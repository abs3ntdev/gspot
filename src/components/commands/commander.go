package commands

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"go.uber.org/fx"
)

type CommanderResult struct {
	fx.Out

	Commander *Commander
}

type CommanderParams struct {
	fx.In

	Context context.Context
	Client  *spotify.Client
}

type Commander struct {
	Context context.Context
	Client  *spotify.Client
}

func NewCommander(p CommanderParams) CommanderResult {
	c := &Commander{
		Context: p.Context,
		Client:  p.Client,
	}
	return CommanderResult{
		Commander: c,
	}
}
