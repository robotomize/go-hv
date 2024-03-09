package gosnap

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robotomize/go-hv/pkg/bash"
	"github.com/robotomize/go-hv/pkg/zsh"
	"golang.org/x/sync/errgroup"
)

const (
	histTypZsh  = "zsh"
	histTypBash = "bash"
)

type Option func(options *Snap)

func WithZSH() Option {
	return func(s *Snap) {
		for _, h := range s.histFiles {
			if strings.Contains(h, histTypZsh) {
				return
			}
		}
		s.histFiles = append(s.histFiles, filepath.Join(s.homePth, ".zsh_history"))
	}
}

func WithBASH() Option {
	return func(s *Snap) {
		for _, h := range s.histFiles {
			if strings.Contains(h, histTypBash) {
				return
			}
		}
		s.histFiles = append(s.histFiles, filepath.Join(s.homePth, ".bash_history"))
	}
}

func New(homePth, histPth string, marshaller Marshaller, opts ...Option) *Snap {
	s := Snap{
		homePth: homePth, histPth: histPth, marshaller: marshaller,
	}
	for _, o := range opts {
		o(&s)
	}

	return &s
}

type Snap struct {
	histFiles  []string
	homePth    string
	histPth    string
	marshaller Marshaller
}

func (s *Snap) Snapshot(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errGrp, ctx := errgroup.WithContext(ctx)

	for _, pth := range s.histFiles {
		pth := pth
		errGrp.Go(
			func() error {
				if err := s.dumpHist(ctx, pth); err != nil {
					return fmt.Errorf("dump %s file: %w", pth, err)
				}
				return nil
			},
		)
	}
	if err := errGrp.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Snap) dumpHist(ctx context.Context, pth string) error {
	var (
		histTyp        string
		scanProviderFn func(r io.Reader, w io.Writer) Parser
	)
	switch {
	case strings.Contains(pth, histTypZsh):
		histTyp = histTypZsh
		scanProviderFn = func(r io.Reader, w io.Writer) Parser {
			return zsh.NewParser(r, w, s.marshaller)
		}
	case strings.Contains(pth, histTypBash):
		histTyp = histTypBash
		scanProviderFn = func(r io.Reader, w io.Writer) Parser {
			return bash.NewParser(r, w, s.marshaller)
		}
	default:
		return fmt.Errorf("unknown history type")
	}

	outputHistFile := filepath.Join(
		s.homePth, s.histPth, fmt.Sprintf("hist-%s.%s.bak", time.Now().Format(time.RFC3339), histTyp),
	)

	outputFile, err := os.OpenFile(outputHistFile, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer outputFile.Close()

	histFile, err := os.Open(pth)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	const flushBuffer = 100 * 1024
	outputWriter := NewSizedBuffer(bufio.NewWriter(outputFile), flushBuffer)
	scanner := scanProviderFn(histFile, outputWriter)
	if err := scanner.Parse(ctx); err != nil {
		return fmt.Errorf("scanner Parse: %w", err)
	}

	if err := outputWriter.Flush(); err != nil {
		return fmt.Errorf("file Sync: %w", err)
	}

	return nil
}
