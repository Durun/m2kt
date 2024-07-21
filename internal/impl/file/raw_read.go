package file

import (
	"context"
	"encoding/json"
	"os"

	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/util/either"
)

func NewRawReader(file *os.File) *RawReader {
	return &RawReader{
		decoder: json.NewDecoder(file),
	}
}

type RawReader struct {
	decoder *json.Decoder
}

func (r *RawReader) DumpVideos(_ context.Context) <-chan either.Either[*youtube.SearchResult] {
	ch := make(chan either.Either[*youtube.SearchResult])

	go func() {
		for r.decoder.More() {
			video := new(youtube.SearchResult)
			if err := r.decoder.Decode(video); err != nil {
				ch <- either.ErrorOf[*youtube.SearchResult](err)
				break
			}

			ch <- either.Of(video)
		}

		close(ch)
	}()

	return ch
}
