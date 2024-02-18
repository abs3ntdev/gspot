package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Status() error {
	state, err := c.Cache.GetOrDo("state", func() (string, error) {
		state, err := c.Client.PlayerState(c.Context)
		if err != nil {
			return "", err
		}
		str, err := c.FormatState(state)
		if err != nil {
			return "", nil
		}
		return str, nil
	}, 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Println(state)
	return nil
}

func (c *Commander) FormatState(state *spotify.PlayerState) (string, error) {
	state.Item.AvailableMarkets = []string{}
	state.Item.Album.AvailableMarkets = []string{}
	out, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return "", err
	}
	return (string(out)), nil
}
