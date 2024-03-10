package fileformat

import (
	"fmt"
	"strings"
	"time"
)

const (
	UnmergedPrefix = "unmerged"
	StagedPrefix   = "staged"
)

const DefaultExt = "bak"

const (
	TypZsh  = "zsh"
	TypBash = "bash"
)

func NewUnmerged(typ string) string {
	return Format{
		Prefix: UnmergedPrefix,
		Time:   time.Now(),
		Typ:    typ,
		Ext:    DefaultExt,
	}.String()
}

func NewStaged(typ string) string {
	return Format{
		Prefix: StagedPrefix,
		Time:   time.Now(),
		Typ:    typ,
		Ext:    DefaultExt,
	}.String()
}

func Parse(s string) (Format, error) {
	var f Format

	if idx := strings.Index(s, "-"); idx > 0 {
		f.Prefix = s[:idx]
	}

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

func New(prefix string, typ string, ext string) Format {
	return Format{Prefix: prefix, Typ: typ, Ext: ext}
}

type Format struct {
	Prefix string
	Time   time.Time
	Typ    string
	Ext    string
}

func (f Format) IsZSH() bool {
	return strings.Contains(f.Prefix, TypZsh)
}

func (f Format) IsBash() bool {
	return strings.Contains(f.Prefix, TypBash)
}

func (f Format) IsUnmerged() bool {
	return strings.Contains(f.Prefix, UnmergedPrefix)
}

func (f Format) IsStaged() bool {
	return strings.Contains(f.Prefix, StagedPrefix)
}

func (f Format) String() string {
	return fmt.Sprintf("%s-%s.%s.%s", f.Prefix, f.Time.Format(time.RFC3339), f.Typ, f.Ext)
}
