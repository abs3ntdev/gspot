package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
)

func playbackCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "play",
			Aliases:  []string{"pl", "start", "s"},
			Usage:    "Plays spotify",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Play(ctx, connect.NewRequest(&gspotv1.PlayRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.PlayResponse])
			}),
		},
		{
			Name:      "playurl",
			Aliases:   []string{"plu"},
			Usage:     "Plays a spotify url",
			ArgsUsage: "url",
			Category:  "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if !cmd.Args().Present() {
					return fmt.Errorf("no url provided")
				}
				if cmd.NArg() > 1 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				resp, err := svc.Client.PlayURL(ctx, connect.NewRequest(&gspotv1.PlayURLRequest{
					Url: cmd.Args().First(),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.PlayURLResponse])
			}),
		},
		{
			Name:     "pause",
			Aliases:  []string{"pa"},
			Usage:    "Pauses spotify",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Pause(ctx, connect.NewRequest(&gspotv1.PauseRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.PauseResponse])
			}),
		},
		{
			Name:     "toggleplay",
			Aliases:  []string{"t"},
			Usage:    "Toggles play/pause",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.TogglePlay(ctx, connect.NewRequest(&gspotv1.TogglePlayRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.TogglePlayResponse])
			}),
		},
		{
			Name:      "next",
			Aliases:   []string{"n", "skip"},
			Usage:     "Skips to the next song",
			ArgsUsage: "amount",
			Category:  "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.NArg() > 1 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				amount := int32(1)
				if cmd.NArg() > 0 {
					amt, err := strconv.Atoi(cmd.Args().First())
					if err != nil {
						return err
					}
					amount = int32(amt)
				}
				resp, err := svc.Client.Next(ctx, connect.NewRequest(&gspotv1.NextRequest{Amount: amount}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.NextResponse])
			}),
		},
		{
			Name:     "previous",
			Aliases:  []string{"b", "prev", "back"},
			Usage:    "Skips to the previous song",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Previous(ctx, connect.NewRequest(&gspotv1.PreviousRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.PreviousResponse])
			}),
		},
		{
			Name:      "setdevice",
			Usage:     "Set the active device",
			ArgsUsage: "<device_id>",
			Category:  "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.NArg() == 0 {
					return fmt.Errorf("no device id provided")
				}
				if cmd.NArg() > 1 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				resp, err := svc.Client.SetDevice(ctx, connect.NewRequest(&gspotv1.SetDeviceRequest{
					DeviceId: cmd.Args().First(),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyDevice)
			}),
		},
		{
			Name:     "repeat",
			Usage:    "Toggle repeat mode",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Repeat(ctx, connect.NewRequest(&gspotv1.RepeatRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyRepeat)
			}),
		},
		{
			Name:     "shuffle",
			Usage:    "Toggle shuffle mode",
			Category: "Playback",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Shuffle(ctx, connect.NewRequest(&gspotv1.ShuffleRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyShuffle)
			}),
		},
		volumeCommand(),
		seekCommand(),
	}
}

func volumeCommand() *cli.Command {
	return &cli.Command{
		Name:     "volume",
		Aliases:  []string{"v"},
		Usage:    "Control the volume",
		Category: "Playback",
		Commands: []*cli.Command{
			{
				Name:      "up",
				Usage:     "Increase the volume (default: 5%)",
				ArgsUsage: "[percent]",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					amt := 5
					if cmd.NArg() == 1 {
						var err error
						amt, err = strconv.Atoi(cmd.Args().First())
						if err != nil {
							return fmt.Errorf("invalid amount: %s", cmd.Args().First())
						}
					}
					resp, err := svc.Client.ChangeVolume(ctx, connect.NewRequest(&gspotv1.ChangeVolumeRequest{Amount: int32(amt)}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyVolume)
				}),
			},
			{
				Name:      "down",
				Aliases:   []string{"dn"},
				Usage:     "Decrease the volume (default: 5%)",
				ArgsUsage: "[percent]",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					amt := 5
					if cmd.NArg() == 1 {
						var err error
						amt, err = strconv.Atoi(cmd.Args().First())
						if err != nil {
							return fmt.Errorf("invalid amount: %s", cmd.Args().First())
						}
					}
					resp, err := svc.Client.ChangeVolume(ctx, connect.NewRequest(&gspotv1.ChangeVolumeRequest{Amount: int32(-amt)}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyVolume)
				}),
			},
			{
				Name:      "set",
				Usage:     "Set the volume to a specific level (0-100)",
				ArgsUsage: "<percent>",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					if cmd.NArg() == 0 {
						return fmt.Errorf("volume level required (0-100)")
					}
					if cmd.NArg() > 1 {
						return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
					}
					vol, err := strconv.Atoi(cmd.Args().First())
					if err != nil {
						return fmt.Errorf("invalid volume: %s", cmd.Args().First())
					}
					if vol < 0 || vol > 100 {
						return fmt.Errorf("volume must be between 0 and 100")
					}
					resp, err := svc.Client.SetVolume(ctx, connect.NewRequest(&gspotv1.SetVolumeRequest{VolumePercent: int32(vol)}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettySetVolume)
				}),
			},
			{
				Name:    "mute",
				Aliases: []string{"m"},
				Usage:   "Mute",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					resp, err := svc.Client.Mute(ctx, connect.NewRequest(&gspotv1.MuteRequest{}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.MuteResponse])
				}),
			},
			{
				Name:    "unmute",
				Aliases: []string{"um"},
				Usage:   "Unmute",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					resp, err := svc.Client.UnMute(ctx, connect.NewRequest(&gspotv1.UnMuteRequest{}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.UnMuteResponse])
				}),
			},
			{
				Name:    "togglemute",
				Aliases: []string{"tm"},
				Usage:   "Toggle mute",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					resp, err := svc.Client.ToggleMute(ctx, connect.NewRequest(&gspotv1.ToggleMuteRequest{}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.ToggleMuteResponse])
				}),
			},
		},
	}
}

func seekCommand() *cli.Command {
	return &cli.Command{
		Name:     "seek",
		Usage:    "Seek to a position in the song",
		Aliases:  []string{"sk"},
		Category: "Playback",
		Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
			if cmd.NArg() > 1 {
				return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
			}
			pos, err := strconv.Atoi(cmd.Args().First())
			if err != nil {
				return err
			}
			resp, err := svc.Client.SetPosition(ctx, connect.NewRequest(&gspotv1.SetPositionRequest{PositionMs: int32(pos)}))
			if err != nil {
				return err
			}
			return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.SetPositionResponse])
		}),
		Commands: []*cli.Command{
			{
				Name:    "forward",
				Aliases: []string{"f"},
				Usage:   "Seek forward",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					resp, err := svc.Client.Seek(ctx, connect.NewRequest(&gspotv1.SeekRequest{Forward: true}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.SeekResponse])
				}),
			},
			{
				Name:    "backward",
				Aliases: []string{"b"},
				Usage:   "Seek backward",
				Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
					resp, err := svc.Client.Seek(ctx, connect.NewRequest(&gspotv1.SeekRequest{Forward: false}))
					if err != nil {
						return err
					}
					return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.SeekResponse])
				}),
			},
		},
	}
}
