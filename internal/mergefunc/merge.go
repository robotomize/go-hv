package mergefunc

import (
	"context"
	"fmt"
	"os"

	"github.com/robotomize/go-hv/internal/fileformat"
)

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

var defaultFilePrefix = ".hist"

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
		case parsed.IsBash():
		default:
			continue
		}
	}

	return nil
}
