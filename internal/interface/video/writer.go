package video

import (
	"context"
	"google.golang.org/api/youtube/v3"
)

type Writer interface {
	WriteVideos(ctx context.Context, videos []*youtube.SearchResult) error
}
