package commands

func (c *Commander) Pause() error {
	return c.Client.Pause(c.Context)
}
