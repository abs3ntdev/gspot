package commands

import (
	"fmt"
	"time"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) NowPlaying(force bool) error {
	if force {
		current, err := c.Client().PlayerCurrentlyPlaying(c.Context)
		if err != nil {
			return err
		}
		str := FormatSong(current)
		fmt.Println(str)
		_, err = c.Cache.Put("now_playing", str, 5*time.Second)
		return err
	}
	song, err := c.Cache.GetOrDo("now_playing", func() (string, error) {
		current, err := c.Client().PlayerCurrentlyPlaying(c.Context)
		if err != nil {
			return "", err
		}
		str := FormatSong(current)
		return str, nil
	}, 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Println(song)
	return nil
}

func FormatSong(current *spotify.CurrentlyPlaying) string {
	out := "▶"
	if !current.Playing || current == nil {
		out = "⏸"
	}
	if current != nil {
		if current.Item != nil {
			out += fmt.Sprintf(" %s", current.Item.Name)
			if len(current.Item.Artists) > 0 {
				out += fmt.Sprintf(" - %s", current.Item.Artists[0].Name)
			}
		}
	}
	return out
}
