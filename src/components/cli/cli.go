package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

var Version = "dev"

func Run(c *commands.Commander, s fx.Shutdowner) {
	defer func() {
		err := s.Shutdown()
		if err != nil {
			c.Log.Error("SHUTDOWN", "error shutting down", err)
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
				Category: "Playback",
			},
			{
				Name:      "playurl",
				Aliases:   []string{"plu"},
				Usage:     "Plays a spotify url",
				Args:      true,
				ArgsUsage: "url",
				Action: func(ctx *cli.Context) error {
					return c.PlayUrl(ctx.Args().First())
				},
				Category: "Playback",
			},
			{
				Name:    "pause",
				Aliases: []string{"pa"},
				Usage:   "Pauses spotify",
				Action: func(cCtx *cli.Context) error {
					return c.Pause()
				},
				Category: "Playback",
			},
			{
				Name:    "toggleplay",
				Aliases: []string{"t"},
				Usage:   "Toggles play/pause",
				Action: func(cCtx *cli.Context) error {
					return c.TogglePlay()
				},
				Category: "Playback",
			},
			{
				Name:    "link",
				Aliases: []string{"yy"},
				Usage:   "Prints the current song's spotify link",
				Action: func(cCtx *cli.Context) error {
					return c.PrintLink()
				},
				Category: "Sharing",
			},
			{
				Name:    "linkcontext",
				Aliases: []string{"lc"},
				Usage:   "Prints the current album or playlist",
				Action: func(cCtx *cli.Context) error {
					return c.PrintLinkContext()
				},
				Category: "Sharing",
			},
			{
				Name:    "youtube-link",
				Aliases: []string{"yl"},
				Usage:   "Prints the current song's youtube link",
				Action: func(cCtx *cli.Context) error {
					return c.PrintYoutubeLink()
				},
				Category: "Sharing",
			},
			{
				Name:    "next",
				Aliases: []string{"n", "skip"},
				Usage:   "Skips to the next song",
				Action: func(cCtx *cli.Context) error {
					return c.Next()
				},
				Category: "Playback",
			},
			{
				Name:    "previous",
				Aliases: []string{"b", "prev", "back"},
				Usage:   "Skips to the previous song",
				Action: func(cCtx *cli.Context) error {
					return c.Previous()
				},
				Category: "Playback",
			},
			{
				Name:    "like",
				Aliases: []string{"l"},
				Usage:   "Likes the current song",
				Action: func(cCtx *cli.Context) error {
					return c.Like()
				},
				Category: "Library Management",
			},
			{
				Name:    "unlike",
				Aliases: []string{"u"},
				Usage:   "Unlikes the current song",
				Action: func(cCtx *cli.Context) error {
					return c.UnLike()
				},
				Category: "Library Management",
			},
			{
				Name:    "nowplaying",
				Aliases: []string{"now"},
				Usage:   "Prints the current song",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "bypass cache",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return c.NowPlaying(cCtx.Bool("force"))
				},
				Category: "Info",
			},
			{
				Name:     "volume",
				Aliases:  []string{"v"},
				Usage:    "Control the volume",
				Category: "Playback",
				Subcommands: []*cli.Command{
					{
						Name:      "up",
						Usage:     "Increase the volume",
						Args:      true,
						ArgsUsage: "percent",
						Action: func(cCtx *cli.Context) error {
							amt, err := strconv.Atoi(cCtx.Args().First())
							if err != nil {
								return err
							}
							return c.ChangeVolume(amt)
						},
					},
					{
						Name:      "down",
						Aliases:   []string{"dn"},
						Usage:     "Decrease the volume",
						Args:      true,
						ArgsUsage: "percent",
						Action: func(cCtx *cli.Context) error {
							amt, err := strconv.Atoi(cCtx.Args().First())
							if err != nil {
								return err
							}
							return c.ChangeVolume(-amt)
						},
					},
					{
						Name:    "mute",
						Aliases: []string{"m"},
						Usage:   "Mute",
						Action: func(cCtx *cli.Context) error {
							return c.Mute()
						},
					},
					{
						Name:    "unmute",
						Aliases: []string{"um"},
						Usage:   "Unmute",
						Action: func(cCtx *cli.Context) error {
							return c.UnMute()
						},
					},
					{
						Name:    "togglemute",
						Aliases: []string{"tm"},
						Usage:   "Toggle mute",
						Action: func(cCtx *cli.Context) error {
							return c.ToggleMute()
						},
					},
				},
			},
			{
				Name:      "download_cover",
				Usage:     "Downloads the cover of the current song",
				Aliases:   []string{"dl"},
				Args:      true,
				ArgsUsage: "path",
				BashComplete: func(cCtx *cli.Context) {
					if cCtx.NArg() > 0 {
						return
					}
				},
				Action: func(cCtx *cli.Context) error {
					return c.DownloadCover(cCtx.Args().First())
				},
				Category: "Info",
			},
			{
				Name:    "radio",
				Usage:   "Starts a radio from the current song",
				Aliases: []string{"r"},
				Action: func(cCtx *cli.Context) error {
					return c.Radio()
				},
				Category: "Radio",
			},
			{
				Name:    "clearradio",
				Usage:   "Clears the radio queue",
				Aliases: []string{"cr"},
				Action: func(cCtx *cli.Context) error {
					return c.ClearRadio()
				},
				Category: "Radio",
			},
			{
				Name:    "refillradio",
				Usage:   "Refills the radio queue with similar songs",
				Aliases: []string{"rr"},
				Action: func(cCtx *cli.Context) error {
					return c.RefillRadio()
				},
				Category: "Radio",
			},
			{
				Name:  "status",
				Usage: "Prints the current status",
				Action: func(ctx *cli.Context) error {
					return c.Status()
				},
				Category: "Info",
			},
			{
				Name:    "devices",
				Usage:   "Lists available devices",
				Aliases: []string{"d"},
				Action: func(cCtx *cli.Context) error {
					return c.ListDevices()
				},
				Category: "Info",
			},
			{
				Name:      "setdevice",
				Usage:     "Set the active device",
				Args:      true,
				ArgsUsage: "<device_id>",
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() == 0 {
						return fmt.Errorf("no device id provided")
					}
					return c.SetDevice(spotify.ID(cCtx.Args().First()))
				},
				Category: "Playback",
			},
			{
				Name:  "repeat",
				Usage: "Toggle repeat mode",
				Action: func(cCtx *cli.Context) error {
					return c.Repeat()
				},
				Category: "Playback",
			},
			{
				Name:  "shuffle",
				Usage: "Toggle shuffle mode",
				Action: func(cCtx *cli.Context) error {
					return c.Shuffle()
				},
				Category: "Playback",
			},
			{
				Name:     "seek",
				Usage:    "Seek to a position in the song",
				Aliases:  []string{"sk"},
				Category: "Playback",
				Action: func(cCtx *cli.Context) error {
					pos, err := strconv.Atoi(cCtx.Args().First())
					if err != nil {
						return err
					}
					return c.SetPosition(pos)
				},
				Subcommands: []*cli.Command{
					{
						Name:    "forward",
						Aliases: []string{"f"},
						Usage:   "Seek forward",
						Action: func(cCtx *cli.Context) error {
							return c.Seek(true)
						},
					},
					{
						Name:    "backward",
						Aliases: []string{"b"},
						Usage:   "Seek backward",
						Action: func(cCtx *cli.Context) error {
							return c.Seek(false)
						},
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		c.Log.Error("COMMANDER", "run error", err)
		os.Exit(1)
	}
}
