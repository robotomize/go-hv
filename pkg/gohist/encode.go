package gohist

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var ErrDecodeDelim = errors.New("delim invalid")

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

var _ Marshaller = (*textMarshaller)(nil)

func NewTextMarshaller() Marshaller {
	return &textMarshaller{}
}

type textMarshaller struct{}

func (textMarshaller) Marshal(ts int64, command string) ([]byte, error) {
	buf := make([]byte, 0, 96)
	buf = append(buf, []byte(strconv.FormatInt(ts, 10))...)
	buf = append(buf, '\t', '\t')
	buf = append(buf, []byte(command)...)
	return buf, nil
}

func (textMarshaller) Unmarshal(b []byte) (ts int64, command string, err error) {
	bufs := bytes.Split(b, []byte{'\t', '\t'})
	if len(bufs) < 2 {
		return 0, "", ErrDecodeDelim
	}

	parseInt, err := strconv.ParseInt(string(bufs[0]), 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("strconv ParseInt: %w", err)
	}

	return parseInt, string(bufs[1]), nil
}
