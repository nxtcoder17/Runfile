package logging

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

type LogLevelStyle int

const (
	LogLevelTrace LogLevelStyle = 0
	LogLevelDebug LogLevelStyle = 1
	LogLevelInfo  LogLevelStyle = 2
	LogLevelWarn  LogLevelStyle = 3
	LogLevelError LogLevelStyle = 4

	LogLevelUnknown LogLevelStyle = 5
)

func ParseLogLevel(level string) int {
	switch level {
	case "trace":
		return 0
	case "debug":
		return 1
	case "info":
		return 2
	case "warn":
		return 3
	case "error":
		return 4
	default:
		return 5
	}
}

type Theme struct {
	TaskPrefixStyle lipgloss.Style
	MessageStyle    lipgloss.Style
	SlogKeyStyle    lipgloss.Style

	// Styles are in order, TRACE,DEBUG,INFO,WARN,ERROR
	LogLevelStyles [6]lipgloss.Style

	LogLevelDebugStyle lipgloss.Style
	LogLevelInfoStyle  lipgloss.Style
	LogLevelErrorStyle lipgloss.Style
}

func DefaultTheme() *Theme {
	if os.Getenv("RUNFILE_THEME") == "light" {
		fg := lipgloss.Color("#4a84ad")
		bg := lipgloss.Color("#f4f7fb")

		style := lipgloss.NewStyle().Background(bg).Foreground(fg)

		return &Theme{
			TaskPrefixStyle: style,
			SlogKeyStyle:    style.Bold(true),
			MessageStyle:    style.UnsetBackground().Foreground(lipgloss.Color("#6d787d")),
			LogLevelStyles: [6]lipgloss.Style{
				style.Foreground(lipgloss.Color("#bdbfbe")).Faint(true),                   // TRACE
				style.Foreground(lipgloss.Color("#bdbfbe")).Faint(true),                   // DEBUG
				style.UnsetBackground().Foreground(lipgloss.Color("#099dd6")),             // INFO
				style.UnsetBackground().Foreground(lipgloss.Color("#d6d609")).Faint(true), // WARN
				style.UnsetBackground().Foreground(lipgloss.Color("#c76975")),             // ERROR
				style.UnsetBackground().Foreground(lipgloss.Color("#d6d609")).Bold(true),  // UNKNOWN
			},
		}
	}

	fg := lipgloss.Color("#9addfc")
	// bg := lipgloss.Color("#172830")

	style := lipgloss.NewStyle().Foreground(fg)

	return &Theme{
		TaskPrefixStyle: style.Faint(true),
		// SlogKeyStyle:    style.UnsetBackground().Foreground(lipgloss.Color("#8f8ad4")),
		// SlogKeyStyle:    style.UnsetBackground().Foreground(lipgloss.Color("#75b9ba")),
		// MessageStyle: style.UnsetBackground().Foreground(lipgloss.Color("#e6e8ed")).Faint(true),
		// MessageStyle: style.UnsetBackground().Foreground(lipgloss.Color("#e6e8ed")).Faint(true),
		MessageStyle: style.UnsetBackground().Foreground(lipgloss.Color("#bdbfbe")),
		// SlogKeyStyle: style.UnsetBackground().Foreground(lipgloss.Color("#2fbaf5")),
		SlogKeyStyle:       style.UnsetBackground().Foreground(lipgloss.Color("#85d3d4")),
		LogLevelDebugStyle: style.Foreground(lipgloss.Color("#bdbfbe")).Faint(true),
		LogLevelInfoStyle:  style.Foreground(lipgloss.Color("#099dd6")),
		LogLevelErrorStyle: style.Foreground(lipgloss.Color("#c76975")),
		LogLevelStyles: [6]lipgloss.Style{
			style.Foreground(lipgloss.Color("#bdbfbe")).Faint(true),                   // TRACE
			style.Foreground(lipgloss.Color("#bdbfbe")).Faint(true),                   // DEBUG
			style.Foreground(lipgloss.Color("#099dd6")),                               // INFO
			style.UnsetBackground().Foreground(lipgloss.Color("#d6d609")).Faint(true), // WARN
			style.Foreground(lipgloss.Color("#c76975")),                               // ERROR
			style.UnsetBackground().Foreground(lipgloss.Color("#d6d609")).Bold(true),  // UNKNOWN
		},
	}
}
