package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/abs3ntdev/gspot/src/components/cache"
	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/abs3ntdev/gspot/src/components/logger"
	"github.com/abs3ntdev/gspot/src/services"
)

// Reusable dependency sets.

// ConfigDeps provides *config.Config and *slog.Logger.
var ConfigDeps = fx.Options(
	services.Config,
	fx.Provide(logger.NewLogger),
)

// DaemonDeps provides the full stack: config, logger, cache, commander, managed context.
var DaemonDeps = fx.Options(
	ConfigDeps,
	fx.Provide(
		managedContext,
		cache.NewCache,
		commands.NewCommander,
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
