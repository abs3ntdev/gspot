package commands

func (c *Commander) Shuffle() error {
	state, err := c.Client().PlayerState(c.Context)
	if err != nil {
		return err
	}
	err = c.Client().Shuffle(c.Context, !state.ShuffleState)
	if err != nil {
		return err
	}
	c.Log.Info("COMMANDER", "shuffle state", !state.ShuffleState)
	return nil
}
