package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/components/daemon"
	"github.com/abs3ntdev/gspot/src/config"
)

func daemonCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "daemon",
			Usage:    "Manage the gspot daemon",
			Category: "Daemon",
			Commands: []*cli.Command{
				{
					Name:  "start",
					Usage: "Start the daemon in the background",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config) error {
								pid, err := daemon.Start(conf)
								if err != nil {
									return err
								}
								fmt.Printf("Daemon started (pid %d)\n", pid)
								return nil
							}),
						)
					},
				},
				{
					Name:  "stop",
					Usage: "Stop the running daemon",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config) error {
								if err := daemon.Stop(conf); err != nil {
									return err
								}
								fmt.Println("Daemon stopped")
								return nil
							}),
						)
					},
				},
				{
					Name:  "status",
					Usage: "Check if the daemon is running",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config) error {
								pid, running := daemon.IsRunning(conf)
								if running {
									fmt.Printf("Daemon is running (pid %d)\n", pid)
								} else {
									fmt.Println("Daemon is not running")
								}
								return nil
							}),
						)
					},
				},
				{
					Name:  "run",
					Usage: "Run the daemon in the foreground",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						// Full fx lifecycle — blocks until signal.
						// DaemonDeps includes lifecycle hooks for PID file, HTTP server, and librespot.
						fx.New(DaemonDeps).Run()
						return nil
					},
				},
				{
					Name:  "restart",
					Usage: "Restart the daemon",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config) error {
								_ = daemon.Stop(conf)
								pid, err := daemon.Start(conf)
								if err != nil {
									return err
								}
								fmt.Printf("Daemon restarted (pid %d)\n", pid)
								return nil
							}),
						)
					},
				},
			},
		},
	}
}
