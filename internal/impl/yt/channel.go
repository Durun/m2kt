package yt

import (
	"log/slog"

	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/util/either"
	"github.com/Durun/m2kt/internal/util/slice"
)

func FetchChannels(service *youtube.Service, channelIDs []string) <-chan either.Either[*youtube.Channel] {
	baseCall := service.Channels.List([]string{
		"id",
		"snippet",
		"contentDetails",
		"statistics",
		"status",
	})

	ch := make(chan either.Either[*youtube.Channel])
	go func() {
		defer close(ch)

		for _, channelIDs := range slice.Chunked(channelIDs, 50) {
			call := *baseCall
			response, err := call.Id(channelIDs...).MaxResults(int64(len(channelIDs))).Do()
			slog.Info("called Channels:list")
			if err != nil {
				ch <- either.ErrorOf[*youtube.Channel](err)
				break
			}

			for _, item := range response.Items {
				ch <- either.Of(item)
			}
		}
	}()

	return ch
}
