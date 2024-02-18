package commands

func (c *Commander) Like() error {
	playing, err := c.Client.PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	return c.Client.AddTracksToLibrary(c.Context, playing.Item.ID)
}
