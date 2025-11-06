package commands

import (
	"context"
	"log/slog"
	"sync"

	"github.com/zmb3/spotify/v2"
	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/components/cache"
	"github.com/abs3ntdev/gspot/src/config"
	"github.com/abs3ntdev/gspot/src/services"
)

type CommanderResult struct {
	fx.Out

	Commander *Commander
}

type CommanderParams struct {
	fx.In

	Context context.Context
	Log     *slog.Logger
	Cache   *cache.Cache
	Config  *config.Config
}

type Commander struct {
	Context context.Context
	User    *spotify.PrivateUser
	Log     *slog.Logger
	Cache   *cache.Cache
	mu      sync.RWMutex
	cl      *spotify.Client
	conf    *config.Config
}

func NewCommander(p CommanderParams) CommanderResult {
	c := &Commander{
		Context: p.Context,
		Log:     p.Log,
		Cache:   p.Cache,
		conf:    p.Config,
	}
	return CommanderResult{
		Commander: c,
	}
}

func (c *Commander) Client() *spotify.Client {
	c.mu.Lock()
	if c.cl == nil {
		c.cl = c.connectClient()
	}
	c.mu.Unlock()
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cl
}

func (c *Commander) connectClient() *spotify.Client {
	client, err := services.GetClient(c.conf)
	if err != nil {
		panic(err)
	}
	currentUser, err := client.CurrentUser(c.Context)
	if err != nil {
		panic(err)
	}
	c.User = currentUser
	return client
}
