package commands

func (c *Commander) PrintLink() (string, error) {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return "", err
	}
	return state.Item.ExternalURLs["spotify"], nil
}

func (c *Commander) PrintLinkContext() (string, error) {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return "", err
	}
	return state.PlaybackContext.ExternalURLs["spotify"], nil
}
