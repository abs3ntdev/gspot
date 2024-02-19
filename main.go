package main

import (
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/app"
	"git.asdf.cafe/abs3nt/gspot/src/components/cache"
	"git.asdf.cafe/abs3nt/gspot/src/components/cli"
	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
)

func main() {
	var s fx.Shutdowner
	app := fx.New(
		fx.Populate(&s),
		app.Config,
		fx.Provide(
			cache.NewCache,
			commands.NewCommander,
		),
		fx.Invoke(
			cli.Run,
		),
	)
	app.Run()
}
