package mergefunc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

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
		listStagedZsh   []string
		listUnmergedZsh []string
		lisStagedBash   []string
		lisUnmergedBash []string
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		parsed, err := fileformat.Parse(entry.Name())
		if err != nil {
			return err
		}

		switch {
		case parsed.IsZSH():
			if parsed.IsUnmerged() {
				listUnmergedZsh = append(listUnmergedZsh, entry.Name())
				continue
			}
			listStagedZsh = append(listStagedZsh, entry.Name())
		case parsed.IsBash():
			if parsed.IsUnmerged() {
				lisUnmergedBash = append(lisUnmergedBash, entry.Name())
				continue
			}
			lisStagedBash = append(lisStagedBash, entry.Name())
		default:
			continue
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

	if len(listUnmergedZsh) >= 2 {
		unionFile := fileformat.NewUnmerged(fileformat.TypZsh)
		create, err := os.Create(filepath.Join(m.pth, unionFile))
		if err != nil {
			return err
		}

		for _, fName := range listUnmergedZsh {
			pth := filepath.Join(m.pth, fName)
			open, err := os.Open(pth)
			if err != nil {
				create.Close()
				return err
			}

			if _, err := io.Copy(create, open); err != nil {
				create.Close()
				return err
			}

			open.Close()

			if err := os.Remove(pth); err != nil {
				create.Close()
				return err
			}
			listUnmergedZsh = []string{unionFile}
		}

		unmergedFileName := filepath.Join(m.pth, listUnmergedZsh[0])
		open, err := os.Open(unmergedFileName)
		if err != nil {
			return err
		}

		_ = open

		slices.SortFunc(listStagedZsh, fortFilesFn)

	}

	return nil
}
