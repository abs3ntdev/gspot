package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	golibrespot "github.com/devgianlu/go-librespot"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/abs3ntdev/gspot/src/components/cache"
	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/abs3ntdev/gspot/src/components/daemon"
	librespotpkg "github.com/abs3ntdev/gspot/src/components/librespot"
	"github.com/abs3ntdev/gspot/src/components/logger"
	"github.com/abs3ntdev/gspot/src/services"
)

// Reusable dependency sets.

// ConfigDeps provides *config.Config and *slog.Logger.
var ConfigDeps = fx.Options(
	services.Config,
	fx.Provide(logger.NewLogger),
)

// DaemonDeps provides the full stack for daemon run:
// config, logger, cache, commander, librespot player, server, HTTP server, PID file.
var DaemonDeps = fx.Options(
	ConfigDeps,
	fx.Provide(
		managedContext,
		cache.NewCache,
		commands.NewCommander,
		// Librespot logger adapter
		func(log *slog.Logger) golibrespot.Logger {
			return librespotpkg.NewSlogAdapter(log)
		},
		// Librespot player (nil if not enabled in config)
		librespotpkg.NewPlayer,
		// ConnectRPC server (receives optional librespot player)
		daemon.NewServer,
	),
	// Lifecycle hooks (order matters: PID file first, then HTTP server)
	fx.Invoke(
		daemon.NewPidFile,
		daemon.NewHTTPServer,
	),
	fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
		l := &fxevent.SlogLogger{Logger: log}
		l.UseLogLevel(slog.LevelDebug)
		return l
	}),
)

// run builds an fx app with the given options and executes all fx.Invoke
// functions. Use for short-lived commands that resolve deps and return.
func run(opts ...fx.Option) error {
	opts = append(opts, fx.NopLogger)
	return fx.New(opts...).Err()
}

// managedContext provides a context.Context cancelled when fx stops.
var ErrContextShutdown = errors.New("shutdown")

func managedContext(
	lc fx.Lifecycle,
	s fx.Shutdowner,
	log *slog.Logger,
) context.Context {
	if log == nil {
		log = slog.Default()
	}
	ctx, cn := context.WithCancelCause(context.Background())
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cn(fmt.Errorf("%w: %w", context.Canceled, ErrContextShutdown))
			return nil
		},
	})
	return ctx
}
