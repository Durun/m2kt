package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/impl/file"
	"github.com/Durun/m2kt/internal/impl/sqlite"
	"github.com/Durun/m2kt/internal/impl/yt"
	"github.com/Durun/m2kt/pkg/chu"
)

func fetchChannelCmd(ctx context.Context, args []string) error {
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	apiKey := cmd.String("key", "", "YouTube Data API v3 key (See: https://console.cloud.google.com/apis/credentials/key )")
	dbFilePath := cmd.String("db", "raw.sqlite", "SQLite3 database file path")
	overwrite := cmd.Bool("overwrite", false, "overwrite existing data (needs more API quota!)")
	limit := cmd.Int("limit", 0, "limit fetch count")
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

	// newline delimited channelIDs
	channelIDs := file.NewLineReader(os.Stdin)
	defer channelIDs.RequestClose()

	queryChannelIDs := make([]string, 0)
	for chunk := range channelIDs.Chunked(1000) {
		channelIDs, err := chunk.Get()
		if err != nil {
			return err
		}

		if *overwrite {
			queryChannelIDs = append(queryChannelIDs, channelIDs...)
		} else {
			ids, err := store.ListChannelIDs(ctx, channelIDs)
			if err != nil {
				return err
			}
			duplicatedIDs := make(map[string]struct{})
			for _, id := range ids {
				duplicatedIDs[id] = struct{}{}
			}

			for _, id := range channelIDs {
				if _, ok := duplicatedIDs[id]; !ok {
					queryChannelIDs = append(queryChannelIDs, id)
				}
			}
		}
	}
	if 0 < *limit && *limit < len(queryChannelIDs) {
		queryChannelIDs = queryChannelIDs[:*limit]
	}

	channels := yt.FetchChannels(service, chu.FromSlice(queryChannelIDs))
	defer channels.RequestClose()

	var fetchCount int
	errs := chu.Chunked(channels, 1000).ForEachCloseOnError(func(channels []*youtube.Channel) error {
		fetchCount += len(channels)
		return store.WriteChannels(ctx, time.Now(), channels)
	})
	if 0 < fetchCount {
		slog.Info("fetched channels into DB", slog.Int("query", len(queryChannelIDs)), slog.Int("fetch", fetchCount))
	}
	if 0 < len(errs) {
		return errors.WithStack(errs[0])
	}

	return nil
}
