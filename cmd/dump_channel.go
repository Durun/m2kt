package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/impl/sqlite"
)

func dumpChannelCmd(ctx context.Context, args []string) error {
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	dbFilePath := cmd.String("db", "raw.sqlite", "SQLite3 database file path")
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
		return err
	}
	defer db.Close()

	store := sqlite.NewRawStore(db)
	channels := store.DumpChannels(ctx)
	defer channels.RequestClose()

	errs := channels.ForEachCloseOnError(func(channel *youtube.Channel) error {
		return errors.WithStack(encoder.Encode(channel))
	})
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}
