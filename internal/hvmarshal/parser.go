package hvmarshal

import (
	"context"
)

type Parser interface {
	Parse(ctx context.Context) error
}
