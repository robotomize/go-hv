package zsh

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

func NewParser(r io.Reader, w io.Writer, outputMarshal Marshaller) *Parser {
	return &Parser{w: w, r: bufio.NewScanner(r), outputMarshal: outputMarshal}
}

type Parser struct {
	r             *bufio.Scanner
	w             io.Writer
	outputMarshal Marshaller
}

func (z *Parser) Parse(ctx context.Context) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	localMarshaller := zshMarshaller{}

	for z.r.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := z.r.Text()
		if strings.HasPrefix(line, ": ") && buf.Len() > 0 {
			scanned := buf.Bytes()

			ts, command, err := localMarshaller.Unmarshal(scanned)
			if err != nil {
				return fmt.Errorf("zsh marshaller Unmarshal: %w", err)
			}

			encBytes, err := z.outputMarshal.Marshal(ts, command)
			if err != nil {
				return fmt.Errorf("output marshaller Marshal: %w", err)
			}

			if _, err := z.w.Write(append(encBytes, '\n')); err != nil {
				return fmt.Errorf("writer Write: %w", err)
			}
			buf.Reset()
			buf.WriteString(line)
			continue
		}

		buf.WriteString(line)
	}

	if buf.Len() > 0 {
		scanned := buf.Bytes()
		ts, command, err := localMarshaller.Unmarshal(scanned)
		if err != nil {
			return err
		}

		encBytes, err := z.outputMarshal.Marshal(ts, command)
		if err != nil {
			return fmt.Errorf("encode func: %w", err)
		}

		if _, err := z.w.Write(encBytes); err != nil {
			return fmt.Errorf("writer Write: %w", err)
		}
		buf.Reset()
	}

	return nil
}
