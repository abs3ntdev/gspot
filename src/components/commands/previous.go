package commands

import (
	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Previous() error {
	err := c.Client().Previous(c.Context)
	if err != nil {
		if IsNoActiveError(err) {
			deviceID, err := c.ActivateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PreviousOpt(c.Context, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}
