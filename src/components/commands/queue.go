package commands

import "github.com/zmb3/spotify/v2"

func (c *Commander) QueueSong(id spotify.ID) error {
	err := c.Client().QueueSong(c.Context, id)
	if err != nil {
		if IsNoActiveError(err) {
			deviceID, err := c.ActivateDevice()
			if err != nil {
				return err
			}
			err = c.Client().QueueSongOpt(c.Context, id, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	return nil
}
