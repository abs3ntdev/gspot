package cli

import (
	"context"
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/zmb3/spotify/v2"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"git.asdf.cafe/abs3nt/gspot/src/components/daemon"
	"git.asdf.cafe/abs3nt/gspot/src/components/tui"
	"git.asdf.cafe/abs3nt/gspot/src/components/tuitview"
)

var Version = "dev"

func Run(c *commands.Commander, s fx.Shutdowner) {
	app := &cli.Command{
		Name:                  "gspot",
		EnableShellCompletion: true,
		Version:               Version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Present() {
				return fmt.Errorf("unknown command: %s", strings.Join(cmd.Args().Slice(), " "))
			}
			return tui.StartTea(c, "main")
		},
		Commands: []*cli.Command{
			{
				Name:    "play",
				Aliases: []string{"pl", "start", "s"},
				Usage:   "Plays spotify",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Play", daemon.PlayArgs{})
				},
				Category: "Playback",
			},
			{
				Name:      "playurl",
				Aliases:   []string{"plu"},
				Usage:     "Plays a spotify url",
				ArgsUsage: "url",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if !cmd.Args().Present() {
						return fmt.Errorf("no url provided")
					}
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("PlayURL", daemon.PlayURLArgs{URL: cmd.Args().First()})
				},
				Category: "Playback",
			},
			{
				Name:    "pause",
				Aliases: []string{"pa"},
				Usage:   "Pauses spotify",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Pause", daemon.PauseArgs{})
				},
				Category: "Playback",
			},
			{
				Name:    "toggleplay",
				Aliases: []string{"t"},
				Usage:   "Toggles play/pause",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("TogglePlay", daemon.TogglePlayArgs{})
				},
				Category: "Playback",
			},
			{
				Name:    "link",
				Aliases: []string{"yy"},
				Usage:   "Prints the current song's spotify link",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("PrintLink", daemon.LinkArgs{})
				},
				Category: "Sharing",
			},
			{
				Name:    "linkcontext",
				Aliases: []string{"lc"},
				Usage:   "Prints the current album or playlist",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("PrintLinkContext", daemon.LinkContextArgs{})
				},
				Category: "Sharing",
			},
			{
				Name:    "youtube-link",
				Aliases: []string{"yl"},
				Usage:   "Prints the current song's youtube link",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("PrintYoutubeLink", daemon.YoutubeLinkArgs{})
				},
				Category: "Sharing",
			},
			{
				Name:      "next",
				Aliases:   []string{"n", "skip"},
				Usage:     "Skips to the next song",
				ArgsUsage: "amount",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					amount := 1
					if cmd.NArg() > 0 {
						amt, err := strconv.Atoi(cmd.Args().First())
						if err != nil {
							return err
						}
						amount = amt
					}
					return sendCommandRPC("Next", daemon.NextArgs{Amount: amount})
				},
				Category: "Playback",
			},
			{
				Name:    "previous",
				Aliases: []string{"b", "prev", "back"},
				Usage:   "Skips to the previous song",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Previous", daemon.PreviousArgs{})
				},
				Category: "Playback",
			},
			{
				Name:    "like",
				Aliases: []string{"l"},
				Usage:   "Likes the current song",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Like", daemon.LikeArgs{})
				},
				Category: "Library Management",
			},
			{
				Name:    "unlike",
				Aliases: []string{"u"},
				Usage:   "Unlikes the current song",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("UnLike", daemon.UnlikeArgs{})
				},
				Category: "Library Management",
			},
			{
				Name:    "nowplaying",
				Aliases: []string{"now"},
				Usage:   "Prints the current song",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "force",
						Aliases:     []string{"f"},
						DefaultText: "false",
						Usage:       "bypass cache",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("NowPlaying", daemon.NowPlayingArgs{Force: cmd.Bool("force")})
				},
				Category: "Info",
			},
			{
				Name:     "volume",
				Aliases:  []string{"v"},
				Usage:    "Control the volume",
				Category: "Playback",
				Commands: []*cli.Command{
					{
						Name:      "up",
						Usage:     "Increase the volume",
						ArgsUsage: "percent",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.NArg() > 1 {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							amt, err := strconv.Atoi(cmd.Args().First())
							if err != nil {
								return err
							}
							return sendCommandRPC("ChangeVolume", daemon.VolumeArgs{Amount: amt})
						},
					},
					{
						Name:      "down",
						Aliases:   []string{"dn"},
						Usage:     "Decrease the volume",
						ArgsUsage: "percent",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.NArg() > 1 {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							amt, err := strconv.Atoi(cmd.Args().First())
							if err != nil {
								return err
							}
							return sendCommandRPC("ChangeVolume", daemon.VolumeArgs{Amount: -amt})
						},
					},
					{
						Name:    "mute",
						Aliases: []string{"m"},
						Usage:   "Mute",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Present() {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							return sendCommandRPC("Mute", daemon.MuteArgs{})
						},
					},
					{
						Name:    "unmute",
						Aliases: []string{"um"},
						Usage:   "Unmute",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Present() {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							return sendCommandRPC("UnMute", daemon.UnmuteArgs{})
						},
					},
					{
						Name:    "togglemute",
						Aliases: []string{"tm"},
						Usage:   "Toggle mute",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Present() {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							return sendCommandRPC("ToggleMute", daemon.ToggleMuteArgs{})
						},
					},
				},
			},
			{
				Name:      "download_cover",
				Usage:     "Downloads the cover of the current song",
				Aliases:   []string{"dl"},
				ArgsUsage: "path",
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					if cmd.NArg() > 0 {
						return
					}
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("DownloadCover", daemon.DownloadCoverArgs{Path: cmd.Args().First()})
				},
				Category: "Info",
			},
			{
				Name:    "radio",
				Usage:   "Starts a radio from the current song",
				Aliases: []string{"r"},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Radio", daemon.RadioArgs{})
				},
				Category: "Radio",
			},
			{
				Name:    "clearradio",
				Usage:   "Clears the radio queue",
				Aliases: []string{"cr"},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("ClearRadio", daemon.ClearRadioArgs{})
				},
				Category: "Radio",
			},
			{
				Name:    "refillradio",
				Usage:   "Refills the radio queue with similar songs",
				Aliases: []string{"rr"},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("RefillRadio", daemon.RefillRadioArgs{})
				},
				Category: "Radio",
			},
			{
				Name:  "status",
				Usage: "Prints the current status",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Status", daemon.StatusArgs{})
				},
				Category: "Info",
			},
			{
				Name:    "devices",
				Usage:   "Lists available devices",
				Aliases: []string{"d"},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("ListDevices", daemon.ListDevicesArgs{})
				},
				Category: "Info",
			},
			{
				Name:      "setdevice",
				Usage:     "Set the active device",
				ArgsUsage: "<device_id>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.NArg() == 0 {
						return fmt.Errorf("no device id provided")
					}
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("SetDevice", daemon.SetDeviceArgs{DeviceID: spotify.ID(cmd.Args().First())})
				},
				Category: "Playback",
			},
			{
				Name:  "repeat",
				Usage: "Toggle repeat mode",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Repeat", daemon.RepeatArgs{})
				},
				Category: "Playback",
			},
			{
				Name:  "shuffle",
				Usage: "Toggle shuffle mode",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return sendCommandRPC("Shuffle", daemon.ShuffleArgs{})
				},
				Category: "Playback",
			},
			{
				Name:  "tui",
				Usage: "Starts the TUI",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					return tui.StartTea(c, "main")
				},
			},
			{
				Name:  "tview",
				Usage: "Starts the TUI using tview (experimental)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					// start tview tui
					return tuitview.TuitView(c)
				},
			},
			{
				Name:     "seek",
				Usage:    "Seek to a position in the song",
				Aliases:  []string{"sk"},
				Category: "Playback",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					pos, err := strconv.Atoi(cmd.Args().First())
					if err != nil {
						return err
					}
					return sendCommandRPC("SetPosition", daemon.SetPositionArgs{Position: pos})
				},
				Commands: []*cli.Command{
					{
						Name:    "forward",
						Aliases: []string{"f"},
						Usage:   "Seek forward",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Present() {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							return sendCommandRPC("Seek", daemon.SeekArgs{Fwd: true})
						},
					},
					{
						Name:    "backward",
						Aliases: []string{"b"},
						Usage:   "Seek backward",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Present() {
								return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
							}
							return sendCommandRPC("Seek", daemon.SeekArgs{Fwd: false})
						},
					},
				},
			},
		},
	}
	if err := app.Run(c.Context, os.Args); err != nil {
		c.Log.Error("COMMANDER", "run error", err)
		s.Shutdown(fx.ExitCode(1))
	}
	s.Shutdown()
}

func sendCommandRPC(method string, args interface{}) error {
	client, err := rpc.Dial("unix", "/tmp/gspot.sock")
	if err != nil {
		return fmt.Errorf("could not connect to daemon: %v", err)
	}
	defer client.Close()

	var reply string
	err = client.Call("Handler."+method, args, &reply)
	if err != nil {
		return fmt.Errorf("error calling %s: %v", method, err)
	}

	if reply != "" {
		fmt.Println(reply)
	}

	return nil
}
