package main

import (
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/app"
	"git.asdf.cafe/abs3nt/gospt-ng/src/components/cli"
	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

func main() {
	var s fx.Shutdowner
	app := fx.New(
		fx.Populate(&s),
		app.Config,
		fx.Provide(
			commands.NewCommander,
		),
		fx.Invoke(
			cli.Run,
		),
	)
	app.Run()
}
