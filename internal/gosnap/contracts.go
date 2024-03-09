package gosnap

import (
	"context"
)

type Parser interface {
	Parse(ctx context.Context) error
}

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}
