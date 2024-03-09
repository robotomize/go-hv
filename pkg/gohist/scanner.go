package gohist

import (
	"context"
)

type Scanner interface {
	Parse(ctx context.Context) error
}
