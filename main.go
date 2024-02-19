package main

import (
	"gfx.cafe/util/go/fxplus"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/components/cache"
	"git.asdf.cafe/abs3nt/gspot/src/components/cli"
	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"git.asdf.cafe/abs3nt/gspot/src/components/logger"
	"git.asdf.cafe/abs3nt/gspot/src/services"
)

func main() {
	var s fx.Shutdowner
	app := fx.New(
		fxplus.WithLogger,
		fx.Populate(&s),
		services.Config,
		fx.Provide(
			fxplus.Context,
			cache.NewCache,
			commands.NewCommander,
			logger.NewLogger,
		),
		fx.Invoke(
			cli.Run,
		),
	)
	app.Run()
}
