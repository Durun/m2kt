package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	cmds := map[string]func(ctx context.Context, args []string) error{
		"fetch-video": fetchVideoCmd,
	}
	cmdNames := make([]string, 0, len(cmds))
	for name := range cmds {
		cmdNames = append(cmdNames, name)
	}
	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <subcommand> [args...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, fmt.Sprintf("subcommand: %s", cmdNames))
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	ctx := context.Background()
	cmdName := os.Args[1]
	cmd, ok := cmds[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", cmdName)
		usage()
		os.Exit(1)
	}

	err := cmd(ctx, os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s error: %+v\n", cmdName, err)
		os.Exit(1)
	}
}
