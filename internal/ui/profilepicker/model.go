// Package profilepicker provides a modal for selecting Docker Compose profiles.
package profilepicker

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// Model is the profile picker state.
type Model struct {
	StackName string
	Profiles  []string
	Selected  map[string]bool
	Cursor    int
	Active    bool
}

// New returns a profile picker model pre-populated with available profiles.
func New(stackName string, profiles []string) Model {
	selected := make(map[string]bool)
	// Default profile is auto-selected if present
	for _, p := range profiles {
		selected[p] = p == "default"
	}
	return Model{
		StackName: stackName,
		Profiles:  profiles,
		Selected:  selected,
		Cursor:    0,
		Active:    true,
	}
}

// Toggle selects/deselects the profile at the current cursor.
func (m *Model) Toggle() {
	if m.Cursor < len(m.Profiles) {
		m.Selected[m.Profiles[m.Cursor]] = !m.Selected[m.Profiles[m.Cursor]]
	}
}

// SelectedProfiles returns the list of profiles that are currently selected.
func (m *Model) SelectedProfiles() []string {
	var out []string
	for _, p := range m.Profiles {
		if m.Selected[p] {
			out = append(out, p)
		}
	}
	return out
}

// Update handles keyboard input for the profile picker.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.Cursor < len(m.Profiles)-1 {
				m.Cursor++
			}
		case "k", "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case " ":
			m.Toggle()
		case "enter":
			m.Active = false
			return m, tea.Quit
		case "esc":
			m.Active = false
		}
	}
	return m, nil
}

// Init implements tea.Model for the profile picker.
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the profile picker modal.
func (m Model) View() string {
	var s []string
	s = append(s, lipgloss.NewStyle().Bold(true).Foreground(theme.ColorPrimary).Render("up: "+m.StackName))
	s = append(s, "\n")

	for i, p := range m.Profiles {
		checked := "[x]"
		if !m.Selected[p] {
			checked = "[ ]"
		}
		row := checked + " " + p
		if i == m.Cursor {
			s = append(s, theme.SelectedRowStyle.Render(row))
		} else {
			s = append(s, theme.NormalStyle.Render(row))
		}
	}

	s = append(s, "\n")
	s = append(s, theme.HelpStyle.Render("space toggle | enter confirm | esc cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

// IsActive returns whether the picker is still active (waiting for input).
func (m Model) IsActive() bool {
	return m.Active
}
