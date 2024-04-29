package sizedbuf

import (
	"bufio"
	"io"
)

func New(writer io.Writer, limit int) *Buffer {
	return &Buffer{
		Writer: bufio.NewWriter(writer),
		limit:  limit,
	}
}

type Buffer struct {
	*bufio.Writer
	size  int
	limit int
}

func (cw *Buffer) Write(b []byte) (int, error) {
	n, err := cw.Writer.Write(b)
	if err != nil {
		return n, err
	}
	cw.size += n
	if cw.size >= cw.limit {
		err = cw.Writer.Flush()
		cw.size = 0
	}
	return n, err
}
