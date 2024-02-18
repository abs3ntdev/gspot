package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
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
