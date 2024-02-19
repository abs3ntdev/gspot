package commands

func (c *Commander) UnLike() error {
	playing, err := c.Client().PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	return c.Client().RemoveTracksFromLibrary(c.Context, playing.Item.ID)
}
