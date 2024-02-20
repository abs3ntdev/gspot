package commands

func (c *Commander) TogglePlay() error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return c.Play()
	}
	if state.Playing {
		return c.Pause()
	}
	return c.Play()
}
