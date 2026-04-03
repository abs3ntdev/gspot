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

func playlistCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "playlists",
			Aliases:  []string{"pls"},
			Usage:    "List your playlists",
			Category: "Playlists",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "limit",
					Usage: "Max playlists to show",
					Value: 50,
				},
				&cli.IntFlag{
					Name:  "offset",
					Usage: "Offset for pagination",
					Value: 0,
				},
			},
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				resp, err := svc.Client.ListPlaylists(ctx, connect.NewRequest(&gspotv1.ListPlaylistsRequest{
					Limit:  int32(cmd.Int("limit")),
					Offset: int32(cmd.Int("offset")),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyPlaylists)
			}),
		},
		{
			Name:      "playlist",
			Aliases:   []string{"pl-info"},
			Usage:     "Show tracks in a playlist",
			ArgsUsage: "<playlist_id>",
			Category:  "Playlists",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.NArg() == 0 {
					return fmt.Errorf("no playlist id provided")
				}
				if cmd.NArg() > 1 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				resp, err := svc.Client.GetPlaylist(ctx, connect.NewRequest(&gspotv1.GetPlaylistRequest{
					PlaylistId: cmd.Args().First(),
				}))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyPlaylist)
			}),
		},
		{
			Name:      "play-playlist",
			Aliases:   []string{"plp"},
			Usage:     "Play a playlist",
			ArgsUsage: "<playlist_id> [track_offset]",
			Category:  "Playlists",
			Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
				if cmd.NArg() == 0 {
					return fmt.Errorf("no playlist id provided")
				}
				if cmd.NArg() > 2 {
					return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
				}
				req := &gspotv1.PlayPlaylistRequest{
					PlaylistId: cmd.Args().First(),
				}
				if cmd.NArg() == 2 {
					offset, err := strconv.Atoi(cmd.Args().Get(1))
					if err != nil {
						return fmt.Errorf("invalid track offset: %w", err)
					}
					req.Offset = int32(offset)
				}
				resp, err := svc.Client.PlayPlaylist(ctx, connect.NewRequest(req))
				if err != nil {
					return err
				}
				return CmdOutput(cmd, resp.Msg, prettyEmpty[*gspotv1.PlayPlaylistResponse])
			}),
		},
	}
}
