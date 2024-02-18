package cli

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

func Run(c *commands.Commander, s fx.Shutdowner) {
	defer func() {
		err := s.Shutdown()
		if err != nil {
			slog.Error("SHUTDOWN", "error shutting down", err)
		}
	}()
	app := &cli.App{
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "play",
				Aliases: []string{"p"},
				Usage:   "Plays spotify",
				Action: func(cCtx *cli.Context) error {
					return c.Play()
				},
			},
			{
				Name:    "pause",
				Aliases: []string{"pa"},
				Usage:   "Pauses spotify",
				Action: func(cCtx *cli.Context) error {
					return c.Pause()
				},
			},
			{
				Name:    "toggleplay",
				Aliases: []string{"t"},
				Usage:   "Toggles play/pause",
				Action: func(cCtx *cli.Context) error {
					return c.TogglePlay()
				},
			},
			{
				Name:    "link",
				Aliases: []string{"l"},
				Action: func(cCtx *cli.Context) error {
					return c.PrintLink()
				},
			},
			{
				Name:    "next",
				Aliases: []string{"n"},
				Action: func(cCtx *cli.Context) error {
					return c.Next()
				},
			},
			{
				Name:    "previous",
				Aliases: []string{"b"},
				Action: func(cCtx *cli.Context) error {
					return c.Previous()
				},
			},
			{
				Name:    "like",
				Aliases: []string{"lk"},
				Action: func(cCtx *cli.Context) error {
					return c.Like()
				},
			},
			{
				Name:    "unlike",
				Aliases: []string{"ul"},
				Action: func(cCtx *cli.Context) error {
					return c.UnLike()
				},
			},
			{
				Name:    "nowplaying",
				Aliases: []string{"np"},
				Action: func(cCtx *cli.Context) error {
					return c.NowPlaying()
				},
			},
			{
				Name:      "download_cover",
				Aliases:   []string{"dc"},
				Args:      true,
				ArgsUsage: "download_cover <path>",
				Action: func(cCtx *cli.Context) error {
					return c.DownloadCover(cCtx.Args().First())
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("COMMANDER", "uh oh", err)
		os.Exit(1)
	}
}
