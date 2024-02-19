package commands

func (c *Commander) Seek(fwd bool) error {
	current, err := c.Client().PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	newPos := current.Progress + 5000
	if !fwd {
		newPos = current.Progress - 5000
	}
	err = c.Client().Seek(c.Context, newPos)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commander) SetPosition(pos int) error {
	err := c.Client().Seek(c.Context, pos)
	if err != nil {
		return err
	}
	return nil
}
