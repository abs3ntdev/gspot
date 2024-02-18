package commands

import (
	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Previous() error {
	err := c.Client.Previous(c.Context)
	if err != nil {
		if isNoActiveError(err) {
			deviceId, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client.PreviousOpt(c.Context, &spotify.PlayOptions{
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
