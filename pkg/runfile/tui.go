package runfile

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	label string
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) View() string {
	// s := lipgloss.NewStyle().Border(lipgloss.BlockBorder(), false, false, false, true)
	// return fmt.Sprintf(
	// 	"%s%s%s",
	// 	m.viewport.View(),
	// 	gap,
	// 	m.textarea.View(),
	// )
	return ""
}
