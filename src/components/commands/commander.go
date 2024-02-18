package commands

import (
	"context"
	"log/slog"
	"os"

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
	Log     *slog.Logger
}

type Commander struct {
	Context context.Context
	Client  *spotify.Client
	User    *spotify.PrivateUser
	Log     *slog.Logger
}

func NewCommander(p CommanderParams) CommanderResult {
	currentUser, err := p.Client.CurrentUser(p.Context)
	if err != nil {
		slog.Error("COMMANDER", "error getting current user", err)
		os.Exit(1)
	}
	c := &Commander{
		Context: p.Context,
		Client:  p.Client,
		User:    currentUser,
		Log:     p.Log,
	}
	return CommanderResult{
		Commander: c,
	}
}
