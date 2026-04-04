package cli

import (
	"context"

	"github.com/urfave/cli/v3"
	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/config"

	"github.com/abs3ntdev/gspot/gen/gspot/v1/gspotv1connect"
)

// Service holds the ConnectRPC client for RPC commands.
type Service struct {
	Client gspotv1connect.GspotServiceClient
}

// rpcCommand wraps a command handler that needs a ConnectRPC client.
// It builds a ConfigDeps fx container, constructs the client (with auto-start),
// and passes ctx, cmd, and *Service to the handler.
func rpcCommand(fn func(context.Context, *cli.Command, *Service) error) func(context.Context, *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return run(
			ConfigDeps,
			fx.Invoke(func(conf *config.Config) error {
				client, err := newClient(conf)
				if err != nil {
					return err
				}
				svc := &Service{Client: client}
				return fn(ctx, cmd, svc)
			}),
		)
	}
}
