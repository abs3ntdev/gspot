package commands

import "fmt"

func (c *Commander) PrintLink() error {
	state, err := c.Client.PlayerState(c.Context)
	if err != nil {
		return err
	}
	fmt.Println(state.Item.ExternalURLs["spotify"])
	return nil
}
