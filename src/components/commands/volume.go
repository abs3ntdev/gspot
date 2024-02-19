package commands

func (c *Commander) ChangeVolume(amount int) error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return err
	}
	newVolume := state.Device.Volume + amount
	if newVolume > 100 {
		newVolume = 100
	}
	if newVolume < 0 {
		newVolume = 0
	}
	return c.Client().Volume(c.Context, newVolume)
}

func (c *Commander) Mute() error {
	return c.ChangeVolume(-100)
}

func (c *Commander) UnMute() error {
	return c.ChangeVolume(100)
}

func (c *Commander) ToggleMute() error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return err
	}
	if state.Device.Volume == 0 {
		return c.ChangeVolume(100)
	}
	return c.ChangeVolume(-100)
}
