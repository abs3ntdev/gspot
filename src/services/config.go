package services

import (
	"git.asdf.cafe/abs3nt/gunner"
	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/config"
)

var Config = fx.Options(
	fx.Provide(
		func() *config.Config {
			c := &config.Config{}
			gunner.LoadApp(c, "gspot")
			return c
		},
	),
)
