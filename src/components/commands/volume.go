package commands

func (c *Commander) ChangeVolume(amount int) error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return err
	}
	newVolume := int(state.Device.Volume) + amount
	if newVolume > 100 {
		newVolume = 100
	}
	if newVolume < 0 {
		newVolume = 0
	}
	return c.Client().Volume(c.Context, newVolume)
}

func (c *Commander) SetVolume(volume int) error {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}
	return c.Client().Volume(c.Context, volume)
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
