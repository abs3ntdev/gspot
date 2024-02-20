package commands

import (
	"strings"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) Next(amt int, inqueue bool) error {
	if inqueue {
		for i := 0; i < amt; i++ {
			err := c.Client().Next(c.Context)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if amt == 1 {
		err := c.Client().Next(c.Context)
		if err != nil {
			if isNoActiveError(err) {
				deviceID, err := c.activateDevice()
				if err != nil {
					return err
				}
				err = c.Client().NextOpt(c.Context, &spotify.PlayOptions{
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
	// found := false
	// playingIndex := 0
	current, err := c.Client().PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	playbackContext := current.PlaybackContext.Type
	switch playbackContext {
	case "playlist":
		found := false
		currentTrackIndex := 0
		page := 1
		for !found {
			playlist, err := c.Client().
				GetPlaylistItems(
					c.Context,
					spotify.ID(strings.Split(string(current.PlaybackContext.URI), ":")[2]),
					spotify.Limit(50),
					spotify.Offset((page-1)*50),
				)
			if err != nil {
				return err
			}
			for idx, track := range playlist.Items {
				if track.Track.Track.ID == current.Item.ID {
					currentTrackIndex = idx + (50 * (page - 1))
					found = true
					break
				}
			}
			page++
		}
		pos := currentTrackIndex + amt
		return c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
			PlaybackContext: &current.PlaybackContext.URI,
			PlaybackOffset: &spotify.PlaybackOffset{
				Position: &pos,
			},
		})
	case "album":
		found := false
		currentTrackIndex := 0
		page := 1
		for !found {
			playlist, err := c.Client().
				GetAlbumTracks(
					c.Context,
					spotify.ID(strings.Split(string(current.PlaybackContext.URI), ":")[2]),
					spotify.Limit(50),
					spotify.Offset((page-1)*50),
				)
			if err != nil {
				return err
			}
			for idx, track := range playlist.Tracks {
				if track.ID == current.Item.ID {
					currentTrackIndex = idx + (50 * (page - 1))
					found = true
					break
				}
			}
			page++
		}
		pos := currentTrackIndex + amt
		return c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
			PlaybackContext: &current.PlaybackContext.URI,
			PlaybackOffset: &spotify.PlaybackOffset{
				Position: &pos,
			},
		})
	default:
		for i := 0; i < amt; i++ {
			err := c.Client().Next(c.Context)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
