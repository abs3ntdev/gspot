package commands

import (
	"fmt"

	"git.asdf.cafe/abs3nt/gspot/src/components/youtube"
)

func (c *Commander) PrintYoutubeLink() error {
	state, err := c.Client.PlayerState(c.Context)
	if err != nil {
		return err
	}
	link := youtube.Search(state.Item.Artists[0].Name + state.Item.Name)
	fmt.Println(link)
	return nil
}
