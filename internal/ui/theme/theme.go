// Package theme defines the lipgloss style constants for the dockpose TUI.
package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorBackground is the main terminal background color.
var ColorBackground = lipgloss.Color("#0d1117")

// ColorSurface is the secondary panel/card background.
var ColorSurface = lipgloss.Color("#161b22")

// ColorBorder is the border color for panels and tables.
var ColorBorder = lipgloss.Color("#30363d")

// ColorText is the default foreground text color.
var ColorText = lipgloss.Color("#c9d1d9")

// ColorMuted is the de-emphasized text color.
var ColorMuted = lipgloss.Color("#8b949e")

// ColorPrimary is the primary brand/accent color (Docker blue).
var ColorPrimary = lipgloss.Color("#2496ed")

// ColorAccent is the secondary accent color for headers.
var ColorAccent = lipgloss.Color("#58a6ff")

// ColorSuccess is the color for healthy/running states.
var ColorSuccess = lipgloss.Color("#3fb950")

// ColorDanger is the color for stopped/error states.
var ColorDanger = lipgloss.Color("#f85149")

// ColorWarning is the color for starting/unhealthy states.
var ColorWarning = lipgloss.Color("#d29922")

// ColorSelectedBg is the background for selected list items.
var ColorSelectedBg = lipgloss.Color("#1f6feb")

// ColorSelectedFg is the foreground for selected list items.
var ColorSelectedFg = lipgloss.Color("#ffffff")

// TitleStyle renders the app title in the header bar.
var TitleStyle = lipgloss.NewStyle().
	Foreground(ColorPrimary).
	Bold(true).
	Padding(0, 1)

// HeaderStyle renders section headers.
var HeaderStyle = lipgloss.NewStyle().
	Foreground(ColorAccent).
	Bold(true).
	Padding(0, 1)

// NormalStyle renders default body text.
var NormalStyle = lipgloss.NewStyle().
	Foreground(ColorText)

// MutedStyle renders de-emphasized text.
var MutedStyle = lipgloss.NewStyle().
	Foreground(ColorMuted)

// SelectedRowStyle renders the currently selected list row.
var SelectedRowStyle = lipgloss.NewStyle().
	Foreground(ColorSelectedFg).
	Background(ColorSelectedBg).
	Bold(true)

// BorderStyle applies a rounded border around panels.
var BorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorBorder)

// TableHeaderStyle renders column headers in stack tables.
var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(ColorAccent).
	Bold(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderBottom(true).
	BorderForeground(ColorBorder).
	Padding(0, 1)

// RunningStyle renders healthy/running status text.
var RunningStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)

// StoppedStyle renders stopped/exited status text.
var StoppedStyle = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)

// UnhealthyStyle renders unhealthy/starting status text.
var UnhealthyStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)

// HelpStyle renders the footer keybind help text.
var HelpStyle = lipgloss.NewStyle().
	Foreground(ColorMuted).
	Padding(0, 1)

// HeaderBarStyle is the persistent top bar shown across all screens.
var HeaderBarStyle = lipgloss.NewStyle().
	Foreground(ColorText).
	Background(ColorSurface).
	Padding(0, 1).
	BorderStyle(lipgloss.NormalBorder()).
	BorderBottom(true).
	BorderForeground(ColorBorder)

// FooterBarStyle is the persistent bottom bar shown across all screens.
var FooterBarStyle = lipgloss.NewStyle().
	Foreground(ColorMuted).
	Background(ColorSurface).
	Padding(0, 1).
	BorderStyle(lipgloss.NormalBorder()).
	BorderTop(true).
	BorderForeground(ColorBorder)

// PanelStyle wraps the body of a screen with a rounded border.
var PanelStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Margin(0, 0)

// ContextChipStyle renders the active Docker context chip.
var ContextChipStyle = lipgloss.NewStyle().
	Foreground(ColorSelectedFg).
	Background(ColorPrimary).
	Bold(true)

// DemoChipStyle renders the "demo" chip in the header when running demo mode.
var DemoChipStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#0d1117")).
	Background(ColorWarning).
	Bold(true)

// ToastStyle renders a transient action result toast.
var ToastStyle = lipgloss.NewStyle().
	Foreground(ColorAccent)

// WarningInlineStyle renders inline warning/filter text.
var WarningInlineStyle = lipgloss.NewStyle().
	Foreground(ColorWarning).
	Bold(true)

// CursorStyle renders the row cursor glyph.
var CursorStyle = lipgloss.NewStyle().
	Foreground(ColorAccent).
	Bold(true)

// StatusStyle returns a styled string for a given container status keyword.
func StatusStyle(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch {
	case strings.Contains(s, "unhealthy"), strings.Contains(s, "restarting"), strings.Contains(s, "starting"), strings.Contains(s, "paused"):
		return UnhealthyStyle.Render(status)
	case strings.Contains(s, "running"), strings.Contains(s, "up"), strings.Contains(s, "healthy"):
		return RunningStyle.Render(status)
	case strings.Contains(s, "exited"), strings.Contains(s, "stopped"), strings.Contains(s, "dead"), strings.Contains(s, "created"):
		return StoppedStyle.Render(status)
	default:
		return NormalStyle.Render(status)
	}
}
