package commands

import (
	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Next() error {
	err := c.Client.Next(c.Context)
	if err != nil {
		if isNoActiveError(err) {
			deviceId, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client.NextOpt(c.Context, &spotify.PlayOptions{
				DeviceID: &deviceId,
			})
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}
