package file

import (
	"bufio"
	"io"

	"github.com/pkg/errors"

	"github.com/Durun/m2kt/internal/util/either"
)

func NewLineReader(r io.Reader) *LineReader {
	return &LineReader{
		r: bufio.NewReader(r),
	}
}

type LineReader struct {
	r *bufio.Reader
}

func (r LineReader) Lines() <-chan either.Either[string] {
	ch := make(chan either.Either[string])

	go func() {
		defer close(ch)

		for {
			line, _, err := r.r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}

				ch <- either.ErrorOf[string](errors.WithStack(err))
				break
			}
			if len(line) == 0 {
				continue
			}

			ch <- either.Of[string](string(line))
		}
	}()

	return ch
}
