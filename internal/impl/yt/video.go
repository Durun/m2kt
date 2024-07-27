package yt

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/pkg/chu"
)

type FetchVideosOptions struct {
	Q          string
	RegionCode string
	EventType  string
	Count      uint
}

func FetchVideos(service *youtube.Service, opt FetchVideosOptions) chu.ReadChan[*youtube.SearchResult] {
	return chu.Generate(func(ctx context.Context, out chu.WriteChan[*youtube.SearchResult]) {
		baseCall := service.Search.List([]string{"id", "snippet"}).
			Type("video").
			EventType(opt.EventType).
			Order("date").
			RegionCode(opt.RegionCode).
			SafeSearch("none").
			Q(opt.Q).
			MaxResults(50)

		call := *baseCall
		count := uint(0)
		pageToken := ""
		for {
			select {
			case <-ctx.Done():
				out.PushError(errors.WithStack(ctx.Err()))
				return
			default:
			}
			if opt.Count <= count {
				break
			}

			call.PageToken(pageToken)
			response, err := call.Do()
			slog.Info("called Search:list")
			if err != nil {
				out.PushError(errors.WithStack(err))
				break
			}
			count += uint(len(response.Items))

			for _, item := range response.Items {
				out.PushValue(item)
			}

			if response.NextPageToken == "" {
				break
			}
			pageToken = response.NextPageToken
		}
	})
}
