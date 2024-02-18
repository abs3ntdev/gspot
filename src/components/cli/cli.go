package cli

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

var Version = "dev"

func Run(c *commands.Commander, s fx.Shutdowner) {
	defer func() {
		err := s.Shutdown()
		if err != nil {
			slog.Error("SHUTDOWN", "error shutting down", err)
		}
	}()
	app := &cli.App{
		EnableBashCompletion: true,
		Version:              Version,
		Commands: []*cli.Command{
			{
				Name:    "play",
				Aliases: []string{"pl", "start", "s"},
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
				Aliases: []string{"yy"},
				Usage:   "Prints the current song's spotify link",
				Action: func(cCtx *cli.Context) error {
					return c.PrintLink()
				},
			},
			{
				Name:    "next",
				Aliases: []string{"n", "skip"},
				Usage:   "Skips to the next song",
				Action: func(cCtx *cli.Context) error {
					return c.Next()
				},
			},
			{
				Name:    "previous",
				Aliases: []string{"b", "prev", "back"},
				Usage:   "Skips to the previous song",
				Action: func(cCtx *cli.Context) error {
					return c.Previous()
				},
			},
			{
				Name:    "like",
				Aliases: []string{"l"},
				Usage:   "Likes the current song",
				Action: func(cCtx *cli.Context) error {
					return c.Like()
				},
			},
			{
				Name:    "unlike",
				Aliases: []string{"u"},
				Usage:   "Unlikes the current song",
				Action: func(cCtx *cli.Context) error {
					return c.UnLike()
				},
			},
			{
				Name:    "nowplaying",
				Aliases: []string{"now"},
				Usage:   "Prints the current song",
				Action: func(cCtx *cli.Context) error {
					return c.NowPlaying()
				},
			},
			{
				Name:      "download_cover",
				Usage:     "Downloads the cover of the current song",
				Aliases:   []string{"dl"},
				Args:      true,
				ArgsUsage: "download_cover <path>",
				Action: func(cCtx *cli.Context) error {
					return c.DownloadCover(cCtx.Args().First())
				},
			},
			{
				Name:    "radio",
				Usage:   "Starts a radio from the current song",
				Aliases: []string{"r"},
				Action: func(cCtx *cli.Context) error {
					return c.Radio()
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("COMMANDER", "run error", err)
		os.Exit(1)
	}
}
