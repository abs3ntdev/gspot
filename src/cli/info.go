package cli

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
)

func infoCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "nowplaying",
			Aliases:  []string{"now"},
			Usage:    "Prints the current song",
			Category: "Info",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "force",
					Aliases:     []string{"f"},
					DefaultText: "false",
					Usage:       "bypass cache",
				},
			},
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.Args().Present() {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				resp, err := svc.Client.NowPlaying(ctx, connect.NewRequest(&gspotv1.NowPlayingRequest{
					Force: cmd.Bool("force"),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyNowPlaying)
			}),
		},
		{
			Name:     "status",
			Usage:    "Prints the current status",
			Category: "Info",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Status(ctx, connect.NewRequest(&gspotv1.StatusRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyStatus)
			}),
		},
		{
			Name:     "devices",
			Aliases:  []string{"d"},
			Usage:    "Lists available devices",
			Category: "Info",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.ListDevices(ctx, connect.NewRequest(&gspotv1.ListDevicesRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyDevices)
			}),
		},
		{
			Name:      "download_cover",
			Aliases:   []string{"dl"},
			Usage:     "Downloads the cover of the current song",
			ArgsUsage: "path",
			Category:  "Info",
			ShellComplete: func(ctx context.Context, cmd *cli.Command) {
				if cmd.NArg() > 0 {
					return
				}
			},
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.NArg() > 1 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				resp, err := svc.Client.DownloadCover(ctx, connect.NewRequest(&gspotv1.DownloadCoverRequest{
					Path: cmd.Args().First(),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyCover)
			}),
		},
	}
}
