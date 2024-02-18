package commands

import (
	"net/url"
	"strings"

	"github.com/zmb3/spotify/v2"
)

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

func (c *Commander) PlayUrl(urlString string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return err
	}
	track_id := strings.Split(url.Path, "/")[2]
	err = c.Client.QueueSong(c.Context, spotify.ID(track_id))
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice(c.Context)
			if err != nil {
				return err
			}
			err = c.Client.QueueSongOpt(c.Context, spotify.ID(track_id), &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			err = c.Client.NextOpt(c.Context, &spotify.PlayOptions{
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
	err = c.Client.Next(c.Context)
	if err != nil {
		return err
	}
	return nil
}
