package mergetool

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/robotomize/go-bf/bf/bfbits"
	"github.com/robotomize/go-hv/internal/fileformat"
	"github.com/robotomize/go-hv/pkg/bash"
	"github.com/robotomize/go-hv/pkg/limitbuf"
	"github.com/robotomize/go-hv/pkg/zsh"
)

type marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

type nextReader interface {
	ReadLine() (ts int64, command string, raw []byte, err error)
	Next() bool
}

func New(pth string) *Mergetool {
	return &Mergetool{pth: pth}
}

type Mergetool struct {
	pth        string
	marshaller marshaller
}

func (m *Mergetool) Merge(ctx context.Context) error {
	if err := m.merge(ctx); err != nil {
		return fmt.Errorf("merge zsh: %w", err)
	}

	return nil
}

func (m *Mergetool) merge(ctx context.Context) error {

	zFile, err := os.OpenFile(filepath.Join(m.pth, ".zsh_history"), os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}

	defer zFile.Close()

	zWriter := limitbuf.New(zFile, 10*1024)
	defer zWriter.Flush()

	bFile, err := os.OpenFile(
		filepath.Join(m.pth, ".bash_history"), os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0644,
	)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}

	defer bFile.Close()

	bWriter := limitbuf.New(zFile, 10*1024)
	defer bWriter.Flush()

	entries, err := os.ReadDir(m.pth)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), "zsh.bak") {
			pth := filepath.Join(m.pth, entry.Name())
			f, err := os.Open(pth)
			if err != nil {
				return fmt.Errorf("os.Open: %w", err)
			}

			keyFn := func(ts int64, cmd string) []byte {
				bb := make([]byte, 0, 64)
				binary.AppendVarint(bb, ts)
				return bb
			}
			filter, err := m.prepareBloomFilter(
				filepath.Join(m.pth, ".zsh_history"), zsh.NewMarshaller(), keyFn,
			)
			if err != nil {
				return fmt.Errorf("prepare zsh bloom filter: %w", err)
			}

			if err := m.mergeWriter(
				zsh.NewReader(f), zWriter, newStorage(filter, keyFn),
			); err != nil {
				return fmt.Errorf("zsh merge writer: %w", err)
			}
		}

		if strings.HasSuffix(entry.Name(), "bash.bak") {
			pth := filepath.Join(m.pth, entry.Name())

			f, err := os.Open(pth)
			if err != nil {
				return fmt.Errorf("os.Open: %w", err)
			}

			keyFn := func(ts int64, cmd string) []byte {
				return []byte(cmd)
			}

			filter, err := m.prepareBloomFilter(
				filepath.Join(m.pth, ".bash_history"), bash.NewMarshaller(), keyFn,
			)
			if err != nil {
				return fmt.Errorf("prepare bash bloom filter: %w", err)
			}

			if err := m.mergeWriter(
				bash.NewReader(f), bWriter, newStorage(filter, keyFn),
			); err != nil {
				return fmt.Errorf("bash merge writer: %w", err)
			}
		}
	}

	return nil
}

func (m *Mergetool) countLines(reader io.Reader) (int64, error) {
	scanner := bufio.NewScanner(reader)
	var loc int64 = 0
	for scanner.Scan() {
		loc++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanner Scan Err: %w", err)
	}

	return loc, nil
}

func (m *Mergetool) provideNextReader(r io.Reader, pth string) (nextReader, error) {
	if idx := strings.LastIndex(pth, "/"); idx > 0 {
		parse, err := fileformat.Parse(pth[idx+1:])
		if err != nil {
			return nil, fmt.Errorf("fileformat.Parse: %w", err)
		}
		switch {
		case parse.IsBash():
			return bash.NewReader(r), nil
		case parse.IsZSH():
			return zsh.NewReader(r), nil
		default:
			return nil, fmt.Errorf("typ not found")
		}
	}
	return nil, fmt.Errorf("can not parse pth")
}

func (m *Mergetool) prepareBloomFilter(
	pth string,
	marsh marshaller,
	fn func(ts int64, cmd string) []byte,
) (bfbits.BloomFilter, error) {
	f, err := os.Open(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return bfbits.NewBloomFilter(0), nil
		}
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	defer f.Close()

	loc, err := m.countLines(f)
	if err != nil {
		return nil, fmt.Errorf("count lines: %w", err)
	}

	bf := bfbits.NewBloomFilter(max(int(loc), 1))

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		ts, cmd, err := marsh.Unmarshal(bytes)
		if err != nil {
			return nil, fmt.Errorf("hv marshaller Unmarshal: %w", err)
		}

		value := fn(ts, cmd)
		if err := bf.Add(value); err != nil {
			return nil, fmt.Errorf("bloom keyFn Add: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner Scan Err: %w", err)
	}

	return bf, nil
}

type flusher interface {
	Write(b []byte) (n int, err error)
	Flush() error
}

type uniqStorage interface {
	Add(ts int64, command string) error
	Contains(ts int64, command string) (bool, error)
}

func (m *Mergetool) mergeWriter(r nextReader, w flusher, s uniqStorage) error {
	for r.Next() {
		ts, command, raw, err := r.ReadLine()
		if err != nil {
			return fmt.Errorf("reader ReadLine: %w", err)
		}

		ok, err := s.Contains(ts, command)
		if err != nil {
			return fmt.Errorf("bloom keyFn Contains: %w", err)
		}
		if !ok {
			if _, err := w.Write(raw); err != nil {
				return fmt.Errorf("sized writer Write: %w", err)
			}

			if err := s.Add(ts, command); err != nil {
				return fmt.Errorf("bloom keyFn Add: %w", err)
			}
		}

	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("sized writer Flush: %w", err)
	}

	return nil
}
