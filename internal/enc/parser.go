package enc

import (
	"context"
)

type Parser interface {
	Parse(ctx context.Context) error
}
