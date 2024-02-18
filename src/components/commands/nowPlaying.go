package commands

import (
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) NowPlaying() error {
	current, err := c.Client.PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	str := FormatSong(current)
	fmt.Println(str)
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
