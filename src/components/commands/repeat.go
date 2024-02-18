package commands

func (c *Commander) Repeat() error {
	state, err := c.Client.PlayerState(c.Context)
	if err != nil {
		return err
	}
	newState := "off"
	if state.RepeatState == "off" {
		newState = "context"
	}
	// spotifyd only supports binary value for repeat, context or off, change when/if spotifyd is better
	err = c.Client.Repeat(c.Context, newState)
	if err != nil {
		return err
	}
	c.Log.Info("COMMANDER", "Repeat set to", newState)
	return nil
}
