package app

import (
	"log/slog"
	"os"

	"gfx.cafe/util/go/fxplus"
	"git.asdf.cafe/abs3nt/gunner"
	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/config"
	"git.asdf.cafe/abs3nt/gspot/src/services"
)

var Services = fx.Options(
	fx.NopLogger,
	fx.Provide(
		func() *slog.Logger {
			return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug.Level(),
				TimeFormat: "[15:04:05.000]",
			}))
		},
		services.NewSpotifyClient,
		fxplus.Context,
	),
)

var Config = fx.Options(
	fx.Provide(
		func() *config.Config {
			c := &config.Config{}
			gunner.LoadApp(c, "gspot")
			return c
		},
	),
	Services,
)
