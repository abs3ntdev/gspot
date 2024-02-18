package app

import (
	"log/slog"
	"os"

	"gfx.cafe/util/go/fxplus"
	"git.asdf.cafe/abs3nt/gunner"
	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/config"
	"git.asdf.cafe/abs3nt/gospt-ng/src/services"
)

var Services = fx.Options(
	fx.NopLogger,
	fx.Provide(
		func() *slog.Logger {
			return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
				AddSource: true,
				Level:     slog.LevelDebug.Level(),
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
			gunner.LoadApp(c, "gospt")
			return c
		},
	),
	Services,
)
