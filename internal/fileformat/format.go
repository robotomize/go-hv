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

	if idx := strings.Index(s, "."); idx > 0 {
		t := s[0:idx]
		parse, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return Format{}, fmt.Errorf("time.Parse: %w", err)
		}
		f.Time = parse

		s = s[idx+1:]
		if idx1 := strings.Index(s, "."); idx1 > 0 {
			f.Typ = s[:idx1]
			f.Ext = s[idx1+1:]
		}
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
	f := Format{Typ: typ, Ext: DefaultExt, Time: time.Now()}

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
