package gosnap

import (
	"bufio"
	"io"
)

func NewSizedBuffer(writer io.Writer, limit int) *SizedBuffer {
	return &SizedBuffer{
		Writer: bufio.NewWriter(writer),
		limit:  limit,
	}
}

type SizedBuffer struct {
	*bufio.Writer
	size  int
	limit int
}

func (cw *SizedBuffer) Write(b []byte) (int, error) {
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
