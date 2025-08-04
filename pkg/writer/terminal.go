package writer

import (
	"os"
	"strings"
)

// IsANSITerminal checks if we're running in a TTY that supports ANSI colors
// This consolidates both TTY detection and ANSI support checking
func IsANSITerminal() bool {
	// Check if stdout/stderr are connected to a terminal
	stdout, _ := os.Stdout.Stat()
	stderr, _ := os.Stderr.Stat()
	isTTY := (stdout.Mode()&os.ModeCharDevice) != 0 || (stderr.Mode()&os.ModeCharDevice) != 0

	if !isTTY {
		return false
	}

	// Check if terminal supports ANSI escape codes
	term := os.Getenv("TERM")
	hasANSI := strings.Contains(term, "xterm") || strings.Contains(term, "screen") || strings.Contains(term, "vt100") || strings.Contains(term, "tmux")

	return hasANSI
}
