package commands

import (
	"context"
	"log/slog"

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
	err := c.Play()
	if err != nil {
		slog.Error("Error playing", err)
	}
	return CommanderResult{
		Commander: c,
	}
}
