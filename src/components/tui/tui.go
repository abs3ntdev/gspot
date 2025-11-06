package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abs3ntdev/gspot/src/components/commands"
)

// StartTea the entry point for the UI. Initializes the model.
func StartTea(cmd *commands.Commander, mode string) error {
	m, err := InitMain(cmd, Mode(mode))
	if err != nil {
		return err
	}
	P = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := P.Run(); err != nil {
		return err
	}
	return nil
}
