package entity

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	VideoID      string
	PublishedAt  time.Time
	Title        string
	Description  string
	ThumbnailURL string

	ETag      string
	ChannelID string
}

func NewVideoFromSearchResult(searchResult *youtube.SearchResult) (Video, error) {
	publishedAt, err := time.Parse(time.RFC3339, searchResult.Snippet.PublishedAt)
	if err != nil {
		return Video{}, errors.WithStack(err)
	}

	return Video{
		VideoID:      searchResult.Id.VideoId,
		PublishedAt:  publishedAt,
		Title:        searchResult.Snippet.Title,
		Description:  searchResult.Snippet.Description,
		ThumbnailURL: searchResult.Snippet.Thumbnails.Default.Url,

		ETag:      searchResult.Etag,
		ChannelID: searchResult.Snippet.ChannelId,
	}, nil
}
