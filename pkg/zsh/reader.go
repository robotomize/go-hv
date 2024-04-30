package zsh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

var b = sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 64)) }}

type Reader interface {
	ReadLine() (ts int64, command string, raw []byte, err error)
	Next() bool
}

func NewReader(r io.Reader) Reader {
	return &reader{r: bufio.NewScanner(r), m: marshaller{}}
}

type reader struct {
	r *bufio.Scanner
	m marshaller
}

func (r *reader) Next() bool {
	return r.r.Scan()
}

func (r *reader) ReadLine() (ts int64, command string, raw []byte, err error) {
	buf := b.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		b.Put(buf)
	}()

	for r.r.Scan() {
		line := r.r.Text()
		if !strings.HasPrefix(line, ": ") && buf.Len() > 0 {
			buf.WriteString(line)
			continue
		}

		buf.WriteString(line)
		scanned := buf.Bytes()

		ts, command, err = r.m.Unmarshal(scanned)
		if err != nil {
			return 0, "", nil, fmt.Errorf("zsh marshaller Unmarshal: %w", err)
		}
		break
	}

	if err := r.r.Err(); err != nil {
		return 0, "", nil, fmt.Errorf("bufio Scan Err: %w", err)
	}

	return ts, command, buf.Bytes(), nil
}
