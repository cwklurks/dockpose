// Package envedit provides a modal for viewing and editing stack .env files.
package envedit

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// Entry represents a single .env line.
type Entry struct {
	Key     string
	Value   string
	Comment string
	Masked  bool
}

// Model is the .env editor state.
type Model struct {
	StackName string
	StackPath string
	Entries   []Entry
	Cursor    int
	Editing   bool
	EditKey   string
	EditValue string
	Revealed  map[string]bool
	Saved     bool
	SaveError string
	Active    bool
}

// New loads the .env file from stackPath and returns an editor model.
func New(stackName, stackPath string) Model {
	envPath := filepath.Join(filepath.Dir(stackPath), ".env")
	entries := parseEnvFile(envPath)
	return Model{
		StackName: stackName,
		StackPath: stackPath,
		Entries:   entries,
		Cursor:    0,
		Revealed:  make(map[string]bool),
		Active:    true,
	}
}

// parseEnvFile reads a .env file and returns a list of entries.
// Values matching sensitive patterns are masked by default.
func parseEnvFile(path string) []Entry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var entries []Entry
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			entries = append(entries, Entry{
				Comment: strings.TrimSpace(line),
				Masked:  false,
			})
			continue
		}
		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, "\"")
		masked := isSensitive(key)
		entries = append(entries, Entry{
			Key:    key,
			Value:  value,
			Masked: masked,
		})
	}
	return entries
}

// isSensitive returns true if the key matches sensitive patterns.
func isSensitive(key string) bool {
	upper := strings.ToUpper(key)
	sensitive := []string{"PASSWORD", "SECRET", "TOKEN", "KEY", "API", "CREDENTIAL"}
	for _, s := range sensitive {
		if strings.Contains(upper, s) {
			return true
		}
	}
	return false
}

// Update handles keyboard input for the .env editor.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	return m, m.handleEnvKey(keyMsg)
}

func (m Model) handleEnvKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		m.handleMoveDown()
	case "k", "up":
		m.handleMoveUp()
	case "r":
		m.handleRevealOne()
	case "R":
		m.handleRevealAll()
	case "enter":
		m.handleEnter()
	case "esc":
		m.handleEsc()
	case "s":
		if !m.Editing {
			m.Save()
		}
	}
	return nil
}

func (m *Model) handleMoveDown() {
	if m.Cursor < len(m.Entries)-1 && !m.Editing {
		m.Cursor++
	}
}

func (m *Model) handleMoveUp() {
	if m.Cursor > 0 && !m.Editing {
		m.Cursor--
	}
}

func (m *Model) handleRevealOne() {
	if m.Cursor < len(m.Entries) && !m.Editing {
		if m.Entries[m.Cursor].Masked {
			m.Revealed[m.Entries[m.Cursor].Key] = true
		}
	}
}

func (m *Model) handleRevealAll() {
	if !m.Editing {
		for i := range m.Entries {
			if m.Entries[i].Masked {
				m.Revealed[m.Entries[i].Key] = true
			}
		}
	}
}

func (m *Model) handleEnter() {
	if !m.Editing && m.Cursor < len(m.Entries) {
		if m.Entries[m.Cursor].Comment == "" {
			m.Editing = true
			m.EditKey = m.Entries[m.Cursor].Key
			m.EditValue = m.Entries[m.Cursor].Value
		}
	}
}

func (m *Model) handleEsc() {
	if m.Editing {
		m.Editing = false
	} else {
		m.Active = false
	}
}

// Save writes the current entries back to the .env file.
func (m *Model) Save() {
	envPath := filepath.Join(filepath.Dir(m.StackPath), ".env")
	var lines []string
	for _, e := range m.Entries {
		if e.Comment != "" {
			lines = append(lines, e.Comment)
		} else {
			lines = append(lines, e.Key+"="+e.Value)
		}
	}
	content := strings.Join(lines, "\n") + "\n"
	tmp, err := os.CreateTemp(filepath.Dir(envPath), ".env.*")
	if err != nil {
		m.SaveError = "create temp: " + err.Error()
		return
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.WriteString(content); err != nil {
		m.SaveError = "write: " + err.Error()
		return
	}
	if err := tmp.Close(); err != nil {
		m.SaveError = "close: " + err.Error()
		return
	}
	if err := os.Rename(tmpName, envPath); err != nil {
		m.SaveError = "rename: " + err.Error()
		return
	}
	m.Saved = true
}

// Init implements tea.Model for the .env editor.
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the .env editor modal.
func (m Model) View() string {
	var s []string
	s = append(s, lipgloss.NewStyle().Bold(true).Foreground(theme.ColorPrimary).Render(".env: "+m.StackName))
	s = append(s, "\n")

	for i, e := range m.Entries {
		if e.Comment != "" {
			s = append(s, theme.MutedStyle.Render(e.Comment))
			continue
		}
		display := e.Value
		if e.Masked && !m.Revealed[e.Key] {
			display = strings.Repeat("*", 8)
		}
		row := e.Key + "=" + display
		if m.Editing && m.EditKey == e.Key {
			row = e.Key + "=[" + m.EditValue + "] (editing)"
		}
		if i == m.Cursor && !m.Editing {
			s = append(s, theme.SelectedRowStyle.Render(row))
		} else {
			s = append(s, theme.NormalStyle.Render(row))
		}
	}

	s = append(s, "\n")
	if m.SaveError != "" {
		s = append(s, theme.StoppedStyle.Render("error: "+m.SaveError))
		s = append(s, "\n")
	} else if m.Saved {
		s = append(s, theme.RunningStyle.Render("saved"))
		s = append(s, "\n")
	}
	s = append(s, theme.HelpStyle.Render("j/k navigate | enter edit | r reveal one | R reveal all | s save | esc cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

// IsActive returns whether the editor is still active.
func (m Model) IsActive() bool {
	return m.Active
}
