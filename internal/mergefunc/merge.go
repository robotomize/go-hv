package mergefunc

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/robotomize/go-hv/internal/fileformat"
)

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

const (
	histTypeZsh  = "zsh"
	histTypeBash = "bash"
)

func New(pth string, marshaller Marshaller) *FileMerge {
	return &FileMerge{pth: pth, marshaller: marshaller}
}

type FileMerge struct {
	pth        string
	marshaller Marshaller
}

func (m *FileMerge) Merge(ctx context.Context) error {
	entries, err := os.ReadDir(m.pth)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	var (
		listZsh []string
		lisBash []string
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.Contains(entry.Name(), histTypeZsh) {
			listZsh = append(listZsh, entry.Name())
		}

		if strings.Contains(entry.Name(), histTypeBash) {
			lisBash = append(lisBash, entry.Name())
		}
	}

	fortFilesFn := func(f, f1 string) int {
		t, _ := fileformat.Parse(f)
		t1, _ := fileformat.Parse(f1)
		if t.Time.Before(t1.Time) {
			return 1
		}

		return -1
	}

	slices.SortFunc(listZsh, fortFilesFn)

	return nil
}
