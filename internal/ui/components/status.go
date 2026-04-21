package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// StatusBadge renders a small status indicator with label.
func StatusBadge(label, status string) string {
	style := statusStyle(status)
	return style.Render(fmt.Sprintf("%s: %s", label, status))
}

// statusStyle returns the appropriate lipgloss style for a status string.
func statusStyle(status string) lipgloss.Style {
	switch status {
	case "running", "up", "healthy":
		return theme.RunningStyle
	case "stopped", "exited", "dead", "created":
		return theme.StoppedStyle
	case "unhealthy", "restarting", "starting", "paused":
		return theme.UnhealthyStyle
	default:
		return theme.NormalStyle
	}
}
