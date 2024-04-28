package fileformat

import (
	"fmt"
	"strings"
	"time"
)

const DefaultExt = "bak"

const (
	TypZsh  = "zsh"
	TypBash = "bash"
)

func Parse(s string) (Format, error) {
	var f Format

	var extIdx, typIdx int
	if idx := strings.LastIndex(s, "."); idx > 0 {
		f.Ext = s[idx+1:]
		extIdx = idx
	}

	if idx := strings.Index(s, "."); idx > 0 {
		f.Typ = s[idx+1 : extIdx]
		typIdx = idx
	}

	if idx := strings.Index(s, "-"); idx > 0 {
		t := s[idx+1 : typIdx]
		parse, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return Format{}, fmt.Errorf("time.Parse: %w", err)
		}
		f.Time = parse
	}

	return f, nil
}

type Option func(*Format)

func WithExt(ext string) Option {
	return func(format *Format) {
		format.Ext = ext
	}
}

func New(typ string, opts ...Option) Format {
	f := Format{Typ: typ, Ext: DefaultExt}

	for _, o := range opts {
		o(&f)
	}

	return f
}

type Format struct {
	Time time.Time
	Typ  string
	Ext  string
}

func (f Format) IsZSH() bool {
	return f.Typ == TypZsh
}

func (f Format) IsBash() bool {
	return f.Typ == TypBash
}

func (f Format) String() string {
	return fmt.Sprintf("%s.%s.%s", f.Time.Format(time.RFC3339), f.Typ, f.Ext)
}
