package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/pkg/errors"

	"github.com/Durun/m2kt/internal/impl/file"
	"github.com/Durun/m2kt/internal/impl/sqlite"
	"github.com/Durun/m2kt/internal/util/either"
)

func dumpIndexedVideoCmd(ctx context.Context, args []string) error {
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	dbFilePath := cmd.String("db", "indexed.sqlite", "SQLite3 database file path")
	where := cmd.String("where", "", "videos WHERE clause")
	orderby := cmd.String("orderby", "", "videos ORDER BY clause")
	err := cmd.Parse(args[1:])
	if err != nil {
		cmd.Usage()
		return err
	}

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	encoder := json.NewEncoder(writer)

	db, err := sqlite.NewSQLiteDB(*dbFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()
	store := sqlite.NewIndexedStore(db)

	for videos := range either.Chunked(store.DumpVideos(ctx, *where, *orderby), 1000) {
		if videos.Err != nil {
			return videos.Err
		}

		if err := file.WriteJSONs(encoder, videos.Value); err != nil {
			return err
		}
	}

	return nil
}
