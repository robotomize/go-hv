package gosnap

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robotomize/go-hv/pkg/gohist/bash"
	"github.com/robotomize/go-hv/pkg/gohist/zsh"
	"golang.org/x/sync/errgroup"
)

const (
	zshLabel  = "zsh"
	bashLabel = "bash"
)

type Option func(options *Options)

type Options struct {
	histFiles []string
}

func WithZSH() Option {
	return func(options *Options) {
		for _, h := range options.histFiles {
			if strings.Contains(h, zshLabel) {
				return
			}
		}
		options.histFiles = append(options.histFiles, ".zsh_history")
	}
}

func WithBASH() Option {
	return func(options *Options) {
		for _, h := range options.histFiles {
			if strings.Contains(h, bashLabel) {
				return
			}
		}
		options.histFiles = append(options.histFiles, ".bash_history")
	}
}

func New(homePth, histPth string, marshaller Marshaller, opts ...Option) *Snap {
	s := &Snap{
		homePth: homePth, histPth: histPth, marshaller: marshaller,
	}
	for _, o := range opts {
		o(&s.opts)
	}

	return s
}

type Snap struct {
	opts       Options
	homePth    string
	histPth    string
	marshaller Marshaller
}

func (s *Snap) Snapshot(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	outputHistFile := filepath.Join(
		s.homePth, s.histPth, fmt.Sprintf("hv-hist.%s.bak", time.Now().Format(time.RFC3339)),
	)
	outputFile, err := os.OpenFile(outputHistFile, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer outputFile.Close()
	errGrp, ctx := errgroup.WithContext(ctx)

	const flushBuffer = 100 * 1024

	for _, hf := range s.opts.histFiles {
		hf := hf
		if strings.Contains(hf, zshLabel) {
			errGrp.Go(
				func() error {
					pth := filepath.Join(s.homePth, hf)
					histFile, err := os.Open(pth)
					if err != nil {
						return fmt.Errorf("os.Open: %w", err)
					}

					outputWriter := NewSizedBuffer(bufio.NewWriter(outputFile), flushBuffer)
					scanner := zsh.NewScanner(histFile, outputWriter, s.marshaller)
					if err := scanner.Parse(ctx); err != nil {
						return fmt.Errorf("scanner Parse: %w", err)
					}
					if err := outputWriter.Flush(); err != nil {
						return fmt.Errorf("file Sync: %w", err)
					}
					return nil
				},
			)

		}

		if strings.Contains(hf, bashLabel) {
			errGrp.Go(
				func() error {
					pth := filepath.Join(s.homePth, hf)
					histFile, err := os.Open(pth)
					if err != nil {
						return fmt.Errorf("os.Open: %w", err)
					}
					outputWriter := NewSizedBuffer(bufio.NewWriter(outputFile), flushBuffer)
					scanner := bash.NewScanner(histFile, outputWriter, s.marshaller)
					if err := scanner.Parse(ctx); err != nil {
						return fmt.Errorf("scanner Parse: %w", err)
					}
					if err := outputWriter.Flush(); err != nil {
						return fmt.Errorf("file Sync: %w", err)
					}
					return nil
				},
			)

		}
	}
	if err := errGrp.Wait(); err != nil {
		return err
	}

	return nil
}
