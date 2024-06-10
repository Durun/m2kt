package file

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
	"os"
)

func NewVideoWriter(file *os.File) *Writer {
	return &Writer{
		file: file,
	}
}

type Writer struct {
	file *os.File
}

func (w *Writer) WriteVideos(ctx context.Context, videos []*youtube.SearchResult) error {
	for _, video := range videos {
		json, err := video.MarshalJSON()
		if err != nil {
			return errors.WithStack(err)
		}

		json = append(json, '\n')

		_, err = w.file.Write(json)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
