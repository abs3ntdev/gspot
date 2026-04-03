package cli

import (
	"context"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
)

func sharingCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "link",
			Aliases:  []string{"yy"},
			Usage:    "Prints the current song's spotify link",
			Category: "Sharing",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.GetLink(ctx, connect.NewRequest(&gspotv1.GetLinkRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyLink)
			}),
		},
		{
			Name:     "linkcontext",
			Aliases:  []string{"lc"},
			Usage:    "Prints the current album or playlist",
			Category: "Sharing",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.GetLinkContext(ctx, connect.NewRequest(&gspotv1.GetLinkContextRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyLinkContext)
			}),
		},
		{
			Name:     "youtube-link",
			Aliases:  []string{"yl"},
			Usage:    "Prints the current song's youtube link",
			Category: "Sharing",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.GetYoutubeLink(ctx, connect.NewRequest(&gspotv1.GetYoutubeLinkRequest{}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyYoutubeLink)
			}),
		},
	}
}
