package main

import (
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gospt-ng/src/app"
	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

func main() {
	fx.New(
		app.Config,
		fx.Invoke(
			commands.NewCommander,
		),
	).Run()
}
