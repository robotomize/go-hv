package zsh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrDecodeDelim    = errors.New("delim invalid")
	ErrParseTimestamp = errors.New("timestamp invalid")
)

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

var _ Marshaller = (*marshaller)(nil)

func NewMarshaller() Marshaller {
	return marshaller{}
}

type marshaller struct{}

func (marshaller) Marshal(ts int64, command string) ([]byte, error) {
	return fmt.Appendf([]byte{}, ": %d:0;%s", ts, command), nil
}

func (marshaller) Unmarshal(b []byte) (int64, string, error) {
	parts := strings.Split(string(b), ";")
	if len(parts) < 2 {
		return 0, "", ErrDecodeDelim
	}

	const fixedLen = 12
	filteredNum := make([]rune, 0, fixedLen)
	for _, r := range parts[0] {
		if unicode.IsDigit(r) {
			filteredNum = append(filteredNum, r)
		}
	}

	number, err := strconv.ParseInt(string(filteredNum[:len(filteredNum)-1]), 10, 64)
	if err != nil {
		return 0, "", ErrParseTimestamp
	}

	return number, parts[1], nil
}
