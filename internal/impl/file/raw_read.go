package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/pkg/chu"
)

func NewRawReader(file *os.File) *RawReader {
	return &RawReader{
		decoder: json.NewDecoder(file),
	}
}

type RawReader struct {
	decoder *json.Decoder
}

func (r *RawReader) DumpVideos(ctx context.Context) chu.ReadChan[*youtube.SearchResult] {
	return chu.GenerateContext(ctx, func(ctx context.Context, out chu.WriteChan[*youtube.SearchResult]) {
		for r.decoder.More() {
			select {
			case <-ctx.Done():
				out.PushError(errors.WithStack(ctx.Err()))
				return
			default:
			}

			video := new(youtube.SearchResult)
			if err := r.decoder.Decode(video); err != nil {
				out.PushError(err)
				return
			}

			out.PushValue(video)
		}
	})
}
