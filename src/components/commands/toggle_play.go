package commands

func (c *Commander) TogglePlay() error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return err
	}
	if state.Playing {
		return c.Client().Pause(c.Context)
	}
	return c.Client().Play(c.Context)
}
