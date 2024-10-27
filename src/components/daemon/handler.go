package daemon

import "git.asdf.cafe/abs3nt/gspot/src/components/commands"

type Handler struct {
	Commander *commands.Commander
}

func (h *Handler) Play(args string, reply *string) error {
	err := h.Commander.Play()
	*reply = "hello fucker"
	return err
}
