package services

import (
	"github.com/abs3ntdev/gunner"
	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/config"
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
