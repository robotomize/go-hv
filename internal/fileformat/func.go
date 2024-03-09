package fileformat

import (
	"fmt"
	"strings"
	"time"
)

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

type Format struct {
	Prefix string
	Time   time.Time
	Typ    string
	Ext    string
}

func (f Format) String() string {
	return fmt.Sprintf("%s-%s.%s.%s", f.Prefix, f.Time.Format(time.RFC3339), f.Typ, f.Ext)
}

func NewFormat(typ string) string {
	return Format{
		Prefix: "hist",
		Time:   time.Now(),
		Typ:    typ,
		Ext:    "bak",
	}.String()
}
