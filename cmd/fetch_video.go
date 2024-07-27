package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/impl/sqlite"
	"github.com/Durun/m2kt/internal/impl/yt"
	"github.com/Durun/m2kt/pkg/chu"
)

func fetchVideoCmd(ctx context.Context, args []string) error {
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	apiKey := cmd.String("key", "", "YouTube Data API v3 key (See: https://console.cloud.google.com/apis/credentials/key )")
	dbFilePath := cmd.String("db", "raw.sqlite", "SQLite3 database file path")
	query := cmd.String("q", "", "search query (See: https://developers.google.com/youtube/v3/docs/search/list )")
	regionCode := cmd.String("region", "JP", "region code")
	count := cmd.Uint("count", 50, "max search count")
	eventType := cmd.String("type", "live", "eventType: [completed, live, upcoming]")
	err := cmd.Parse(args[1:])
	if err == nil {
		names := make([]string, 0)
		if *apiKey == "" {
			names = append(names, "key")
		}

		if 0 < len(names) {
			err = fmt.Errorf("required flag: %s", names)
		}
	}
	if err != nil {
		cmd.Usage()
		return err
	}

	db, err := sqlite.NewSQLiteDB(*dbFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()

	store := sqlite.NewRawStore(db)
	if err := store.Prepare(ctx); err != nil {
		return errors.WithStack(err)
	}

	service, err := youtube.NewService(ctx,
		option.WithAPIKey(*apiKey),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	videoCh := yt.FetchVideos(service, yt.FetchVideosOptions{
		Q:          *query,
		RegionCode: *regionCode,
		EventType:  *eventType,
		Count:      *count,
	})
	defer videoCh.RequestClose()
	errs := chu.Chunked(videoCh, 100).ForEachCloseOnError(func(videos []*youtube.SearchResult) error {
		videoIDs := make([]string, 0, len(videos))
		for _, video := range videos {
			videoIDs = append(videoIDs, video.Id.VideoId)
		}
		updateCount, err := store.CountVideos(ctx, videoIDs)
		if err != nil {
			return errors.WithStack(err)
		}
		insertCount := len(videos) - updateCount

		if err := store.WriteVideos(ctx, videos); err != nil {
			return errors.WithStack(err)
		}

		slog.Info("inserted videos into DB", slog.Int("insert", insertCount), slog.Int("update", updateCount))
		if 0 < updateCount {
			videoCh.RequestClose()
		}

		return nil
	})
	if 0 < len(errs) {
		return errors.WithStack(errs[0])
	}

	return nil
}
