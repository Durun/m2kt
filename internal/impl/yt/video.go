package yt

import (
	"log/slog"

	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/util/either"
)

type FetchVideosOptions struct {
	Q          string
	RegionCode string
	EventType  string
	Count      uint
}

func FetchVideos(service *youtube.Service, opt FetchVideosOptions) <-chan either.Either[*youtube.SearchResult] {
	baseCall := service.Search.List([]string{"id", "snippet"}).
		Type("video").
		EventType(opt.EventType).
		Order("date").
		RegionCode(opt.RegionCode).
		SafeSearch("none").
		Q(opt.Q).
		MaxResults(50)

	ch := make(chan either.Either[*youtube.SearchResult])
	count := uint(0)
	go func() {
		defer close(ch)

		call := *baseCall
		pageToken := ""
		for {
			if opt.Count <= count {
				break
			}

			call.PageToken(pageToken)
			response, err := call.Do()
			slog.Info("called Search:list")
			if err != nil {
				ch <- either.ErrorOf[*youtube.SearchResult](err)
				break
			}
			count += uint(len(response.Items))

			for _, item := range response.Items {
				ch <- either.Of(item)
			}

			if response.NextPageToken == "" {
				break
			}
			pageToken = response.NextPageToken
		}
	}()

	return ch
}
