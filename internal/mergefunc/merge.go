package mergefunc

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
	"github.com/robotomize/go-hv/pkg/bash"
	"github.com/robotomize/go-hv/pkg/sizedbuf"
	"github.com/robotomize/go-hv/pkg/zsh"
)

type marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

type nextReader interface {
	ReadLine() (ts int64, command string, err error)
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
	zshFilterKeyFn := func(ts int64, cmd string) []byte {
		bb := make([]byte, 0, 64)
		binary.AppendVarint(bb, ts)
		return bb
	}
	zshFilter, err := m.prepareBloomFilter(
		filepath.Join(m.pth, ".zsh_history"), zsh.NewMarshaller(), zshFilterKeyFn,
	)
	if err != nil {
		return fmt.Errorf("prepare bloom keyFn: %w", err)
	}

	bashFilterKeyFn := func(ts int64, cmd string) []byte {
		return []byte(cmd)
	}

	bashFilter, err := m.prepareBloomFilter(
		filepath.Join(m.pth, ".bash_history"), zsh.NewMarshaller(), bashFilterKeyFn,
	)
	if err != nil {
		return fmt.Errorf("prepare bloom keyFn: %w", err)
	}

	zshFile, err := os.OpenFile(filepath.Join(m.pth, ".zsh_history"), os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}

	defer zshFile.Close()

	zshWriter := sizedbuf.New(zshFile, 10*1024)
	defer zshWriter.Flush()

	bashFile, err := os.OpenFile(
		filepath.Join(m.pth, ".bash_history"), os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0644,
	)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}

	defer bashFile.Close()

	bashWriter := sizedbuf.New(zshFile, 10*1024)
	defer bashWriter.Flush()

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
			if err := m.mergeWriter(
				pth, zshWriter, bloomOpts{
					keyFn: zshFilterKeyFn,
					bloom: zshFilter,
				},
			); err != nil {
				return fmt.Errorf("merge writer: %w", err)
			}
		}

		if strings.HasSuffix(entry.Name(), "bash.bak") {
			pth := filepath.Join(m.pth, entry.Name())
			if err := m.mergeWriter(
				pth, bashWriter, bloomOpts{
					keyFn: bashFilterKeyFn,
					bloom: bashFilter,
				},
			); err != nil {
				return fmt.Errorf("merge writer: %w", err)
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

	bf := bfbits.NewBloomFilter(int(loc))

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

type bloomOpts struct {
	keyFn func(int64, string) []byte
	bloom bfbits.BloomFilter
}

func (m *Mergetool) mergeWriter(pth string, w flusher, bloom bloomOpts) error {
	f, err := os.Open(pth)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	reader := bash.NewReader(f)
	for reader.Next() {
		ts, command, err := reader.ReadLine()
		if err != nil {
			return fmt.Errorf("reader ReadLine: %w", err)
		}

		value := bloom.keyFn(ts, command)
		ok, err := bloom.bloom.Contains(value)
		if err != nil {
			return fmt.Errorf("bloom keyFn Contains: %w", err)
		}
		if !ok {
			if _, err := w.Write(value); err != nil {
				return fmt.Errorf("sized writer Write: %w", err)
			}
		}
		if err := bloom.bloom.Add(value); err != nil {
			return fmt.Errorf("bloom keyFn Add: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("sized writer Flush: %w", err)
	}

	return nil
}
