package commands

import (
	"net/url"
	"strings"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Play() error {
	err := c.Client().Play(c.Context)
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
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

func (c *Commander) PlayLikedSongs(position int) error {
	c.Log.Debug("Playing liked songs")
	err := c.ClearRadio()
	if err != nil {
		return err
	}
	playlist, _, err := c.GetRadioPlaylist("Saved Songs")
	if err != nil {
		return err
	}
	c.Log.Debug("got playlist", "id", playlist.ID)
	songs, err := c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset(position))
	if err != nil {
		return err
	}
	to_add := []spotify.ID{}
	for _, song := range songs.Tracks {
		to_add = append(to_add, song.ID)
	}
	_, err = c.Client().AddTracksToPlaylist(c.Context, playlist.ID, to_add...)
	if err != nil {
		return err
	}
	c.Log.Debug("added songs to playlist")
	err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &playlist.URI,
	})
	if err != nil {
		if isNoActiveError(err) {
			c.Log.Debug("need to activate device")
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
				PlaybackContext: &playlist.URI,
				DeviceID:        &deviceID,
			})
			if err != nil {
				return err
			}
		}
	}
	c.Log.Debug("starting loop")
	for page := 2; page <= 5; page++ {
		c.Log.Debug("doing loop", "page", page)
		songs, err := c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset((50*(page-1))+position))
		if err != nil {
			return err
		}
		to_add := []spotify.ID{}
		for _, song := range songs.Tracks {
			to_add = append(to_add, song.ID)
		}
		_, err = c.Client().AddTracksToPlaylist(c.Context, playlist.ID, to_add...)
		if err != nil {
			return err
		}
	}
	c.Log.Debug("done")
	return err
}

func (c *Commander) PlayUrl(urlString string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return err
	}
	track_id := strings.Split(url.Path, "/")[2]
	err = c.Client().QueueSong(c.Context, spotify.ID(track_id))
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().QueueSongOpt(c.Context, spotify.ID(track_id), &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			err = c.Client().NextOpt(c.Context, &spotify.PlayOptions{
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
	err = c.Client().Next(c.Context)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commander) PlaySongInPlaylist(context *spotify.URI, offset *int) error {
	e := c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
		PlaybackContext: context,
	})
	if e != nil {
		if isNoActiveError(e) {
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
				PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
				PlaybackContext: context,
				DeviceID:        &deviceID,
			})
			if err != nil {
				if isNoActiveError(err) {
					deviceID, err := c.activateDevice()
					if err != nil {
						return err
					}
					err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
						PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
						PlaybackContext: context,
						DeviceID:        &deviceID,
					})
					if err != nil {
						return err
					}
				}
			}
			err = c.Client().Play(c.Context)
			if err != nil {
				return err
			}
		} else {
			return e
		}
	}
	return nil
}
