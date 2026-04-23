// Package contextpicker provides a modal for switching Docker contexts.
package contextpicker

import (
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// Context represents a Docker context with name and endpoint.
type Context struct {
	Name     string
	Endpoint string
	IsLocal  bool
}

// Model is the context picker state.
type Model struct {
	Contexts []Context
	Cursor   int
	Active   bool
	Selected string
	Error    string
}

// New returns a context picker populated with available Docker contexts.
func New() Model {
	contexts := listDockerContexts()
	return Model{
		Contexts: contexts,
		Cursor:   0,
		Active:   true,
	}
}

// listDockerContexts discovers Docker contexts via `docker context ls`.
func listDockerContexts() []Context {
	cmd := exec.Command("docker", "context", "ls", "--format", "{{.Name}}\t{{.Endpoint}}")
	out, err := cmd.Output()
	if err != nil {
		return []Context{{Name: "local", Endpoint: "", IsLocal: true}}
	}
	lines := strings.Split(string(out), "\n")
	var contexts []Context
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		endpoint := strings.TrimSpace(parts[1])
		isLocal := name == "default" || endpoint == "" || strings.HasPrefix(endpoint, "unix://")
		contexts = append(contexts, Context{
			Name:     name,
			Endpoint: endpoint,
			IsLocal:  isLocal,
		})
	}
	return contexts
}

// Update handles keyboard input for the context picker.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.Cursor < len(m.Contexts)-1 {
				m.Cursor++
			}
		case "k", "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "enter":
			if len(m.Contexts) == 0 {
				m.Error = "no docker contexts found"
				return m, nil
			}
			m.Selected = m.Contexts[m.Cursor].Name
			m.Active = false
			return m, nil
		case "esc":
			m.Active = false
		}
	}
	return m, nil
}

// Init implements tea.Model for the context picker.
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the context picker modal.
func (m Model) View() string {
	var s []string
	s = append(s, lipgloss.NewStyle().Bold(true).Foreground(theme.ColorPrimary).Render("contexts"))
	s = append(s, "\n")

	for i, ctx := range m.Contexts {
		marker := " "
		if ctx.Name == "default" || (i == 0 && ctx.IsLocal) {
			marker = "●"
		}
		row := marker + " " + ctx.Name + "  " + ctx.Endpoint
		if i == m.Cursor {
			s = append(s, theme.SelectedRowStyle.Render(row))
		} else {
			s = append(s, theme.NormalStyle.Render(row))
		}
	}

	s = append(s, "\n")
	if m.Error != "" {
		s = append(s, theme.StoppedStyle.Render("error: "+m.Error))
		s = append(s, "\n")
	}
	s = append(s, theme.HelpStyle.Render("j/k navigate | enter switch | esc cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

// IsActive returns whether the picker is still active.
func (m Model) IsActive() bool {
	return m.Active
}

// SelectedContext returns the name of the selected context.
func (m Model) SelectedContext() string {
	return m.Selected
}
