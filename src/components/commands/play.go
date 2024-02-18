package commands

import "github.com/zmb3/spotify/v2"

func (c *Commander) Play() error {
	err := c.Client.Play(c.Context)
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice(c.Context)
			if err != nil {
				return err
			}
			err = c.Client.PlayOpt(c.Context, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
