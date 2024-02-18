package commands

func (c *Commander) Play() error {
	return c.Client.Play(c.Context)
}
