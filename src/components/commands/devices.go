package commands

import (
	"encoding/json"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) ListDevices() error {
	devices, err := c.Client.PlayerDevices(c.Context)
	if err != nil {
		return err
	}
	return PrintDevices(devices)
}

func PrintDevices(devices []spotify.PlayerDevice) error {
	out, err := json.MarshalIndent(devices, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
