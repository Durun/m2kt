package file

import (
	"bufio"
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/Durun/m2kt/pkg/chu"
)

func NewLineReader(r io.Reader) chu.ReadChan[string] {
	return chu.Generate(func(ctx context.Context, out chu.WriteChan[string]) {
		buf := bufio.NewReader(r)
		for {
			select {
			case <-ctx.Done():
				out.PushError(errors.WithStack(ctx.Err()))
				return
			default:
			}

			line, _, err := buf.ReadLine()
			switch {
			case err == io.EOF:
				return
			case err != nil:
				out.PushError(errors.WithStack(err))
				return
			case len(line) == 0:
				continue
			}

			out.PushValue(string(line))
		}
	})
}
