package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var Version = "dev"

func Run() {
	app := &cli.Command{
		Name:                  "gspot",
		EnableShellCompletion: true,
		Version:               Version,
		Usage:                 "A Spotify CLI controller",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format: pretty, json, silent",
				Value: "pretty",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Shorthand for --format=json",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Write output to file instead of stdout (use - for stdout)",
			},
		},
		Commands: allCommands(),
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func allCommands() []*cli.Command {
	var cmds []*cli.Command
	cmds = append(cmds, playbackCommands()...)
	cmds = append(cmds, sharingCommands()...)
	cmds = append(cmds, infoCommands()...)
	cmds = append(cmds, libraryCommands()...)
	cmds = append(cmds, playlistCommands()...)
	cmds = append(cmds, librespotCommands()...)
	cmds = append(cmds, daemonCommands()...)
	return cmds
}
