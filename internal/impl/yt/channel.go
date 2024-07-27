package yt

import (
	"log/slog"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/pkg/chu"
)

func FetchChannels(service *youtube.Service, channelIDs chu.ReadChan[string]) chu.ReadChan[*youtube.Channel] {
	baseCall := service.Channels.List([]string{
		"id",
		"snippet",
		"contentDetails",
		"statistics",
		"status",
	})

	chunks := chu.Map(chu.Chunked(channelIDs, 50), func(channelIDs []string) ([]*youtube.Channel, error) {
		call := *baseCall
		response, err := call.Id(channelIDs...).MaxResults(int64(len(channelIDs))).Do()
		slog.Info("called Channels:list")
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return response.Items, nil
	})

	return chu.Flatten(chunks)
}
