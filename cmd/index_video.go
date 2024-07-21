package main

import (
	"context"
	"flag"
	"os"

	"github.com/pkg/errors"

	"github.com/Durun/m2kt/internal/entity"
	"github.com/Durun/m2kt/internal/impl/file"
	"github.com/Durun/m2kt/internal/impl/sqlite"
	"github.com/Durun/m2kt/internal/util/either"
)

func indexVideoCmd(ctx context.Context, args []string) error {
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	dbFilePath := cmd.String("db", "indexed.sqlite", "SQLite3 database file path")
	err := cmd.Parse(args[1:])
	if err != nil {
		cmd.Usage()
		return err
	}

	reader := file.NewRawReader(os.Stdin)

	db, err := sqlite.NewSQLiteDB(*dbFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()
	store := sqlite.NewIndexedStore(db)

	for rawVideos := range either.Chunked(reader.DumpVideos(ctx), 1000) {
		if rawVideos.Err != nil {
			return errors.WithStack(rawVideos.Err)
		}

		videos := make([]entity.Video, 0, len(rawVideos.Value))
		for _, rawVideo := range rawVideos.Value {
			video, err := entity.NewVideoFromSearchResult(rawVideo)
			if err != nil {
				return errors.WithStack(err)
			}

			videos = append(videos, video)
		}

		if err := store.WriteVideos(ctx, videos); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
