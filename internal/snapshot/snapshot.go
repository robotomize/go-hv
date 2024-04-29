package snapshot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/robotomize/go-hv/internal/fileformat"
	"golang.org/x/sync/errgroup"
)

const (
	histTypZsh  = "zsh"
	histTypBash = "bash"
)

var ErrFileNotFound = errors.New("file not found")

type Option func(options *Snap)

func New(homePth, histPth string, marshaller Marshaller, opts ...Option) *Snap {
	s := Snap{
		homePth: homePth, histPth: histPth, marshaller: marshaller, histFiles: []string{
			filepath.Join(homePth, ".bash_history"),
			filepath.Join(homePth, ".zsh_history"),
		},
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
				if err := s.dump(pth); err != nil {
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

func (s *Snap) dump(pth string) error {
	var histTyp string

	switch {
	case strings.Contains(pth, histTypZsh):
		histTyp = histTypZsh
	case strings.Contains(pth, histTypBash):
		histTyp = histTypBash
	default:
		return fmt.Errorf("unknown history type")
	}

	input, err := os.Open(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrFileNotFound
		}
		return fmt.Errorf("os.Open: %w", err)
	}

	outputFName := filepath.Join(s.homePth, s.histPth, fileformat.New(histTyp).String())
	output, err := os.OpenFile(outputFName, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}
