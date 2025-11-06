package commands

import (
	"github.com/abs3ntdev/gspot/src/components/youtube"
)

func (c *Commander) PrintYoutubeLink() (string, error) {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return "", err
	}
	link := youtube.Search(state.Item.Artists[0].Name + state.Item.Name)
	return link, nil
}
