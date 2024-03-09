package bash

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

func NewParser(r io.Reader, w io.Writer, outputMarshaller Marshaller) *Parser {
	return &Parser{w: w, r: bufio.NewScanner(r), marshaller: outputMarshaller}
}

type Parser struct {
	r          *bufio.Scanner
	w          io.Writer
	marshaller Marshaller
}

func (z *Parser) Parse(ctx context.Context) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	localMarshaller := marshaller{}
	for z.r.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := z.r.Text()
		if strings.HasSuffix(line, "\\") && buf.Len() > 0 {
			buf.WriteString(line)
			continue
		}

		buf.WriteString(line)

		scanned := buf.Bytes()
		ts, command, err := localMarshaller.Unmarshal(scanned)
		if err != nil {
			return fmt.Errorf("bash marshaller Unmarshal: %w", err)
		}

		encBytes, err := z.marshaller.Marshal(ts, command)
		if err != nil {
			return fmt.Errorf("output marshaller Marshal: %w", err)
		}

		if _, err := z.w.Write(append(encBytes, '\n')); err != nil {
			return fmt.Errorf("writer Write: %w", err)
		}
		buf.Reset()
	}

	return nil
}
