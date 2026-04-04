package cli

import (
	"context"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
)

func libraryCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "like",
			Aliases:  []string{"l"},
			Usage:    "Likes the current song",
			Category: "Library",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.Like(ctx, connect.NewRequest(&gspotv1.LikeRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.LikeResponse])
			}),
		},
		{
			Name:     "unlike",
			Aliases:  []string{"u"},
			Usage:    "Unlikes the current song",
			Category: "Library",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.UnLike(ctx, connect.NewRequest(&gspotv1.UnLikeRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.UnLikeResponse])
			}),
		},
	}
}
