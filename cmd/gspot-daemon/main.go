package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"git.asdf.cafe/abs3nt/gspot/src/components/cache"
	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"git.asdf.cafe/abs3nt/gspot/src/components/daemon"
	"git.asdf.cafe/abs3nt/gspot/src/components/logger"
	"git.asdf.cafe/abs3nt/gspot/src/services"
)

func main() {
	var s fx.Shutdowner
	app := fx.New(
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			l := &fxevent.SlogLogger{Logger: logger}
			l.UseLogLevel(slog.LevelDebug)
			return l
		}),
		fx.Populate(&s),
		services.Config,
		fx.Provide(
			Context,
			cache.NewCache,
			commands.NewCommander,
			logger.NewLogger,
		),
		fx.Invoke(
			daemon.Run,
		),
	)
	app.Run()
}

type AsyncInit func(func(ctx context.Context) error)

var ErrContextShutdown = errors.New("shutdown")

func Context(
	lc fx.Lifecycle,
	s fx.Shutdowner,
	log *slog.Logger,
) (context.Context, AsyncInit) {
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
	return ctx, func(fn func(ctx context.Context) error) {
		go func() {
			err := fn(ctx)
			if err != nil {
				log.Error("Failed to run async hook", "err", err)
				s.Shutdown()
			}
		}()
	}
}
