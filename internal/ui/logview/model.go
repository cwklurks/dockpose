// Package logview provides a streaming log viewer bubble tea component.
package logview

import (
	"context"
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// ServiceLog ties a service name to its log channel.
type ServiceLog struct {
	ServiceName string
	ContainerID string
	Lines       <-chan string
}

// LogModel is the bubble Tea model for the streaming log viewer.
type LogModel struct {
	Services   []ServiceLog
	Buffer     []string
	Follow     bool
	Timestamps bool
	Wrap       bool
	FilterText string
	Cursor     int
	width      int

	cancelFn context.CancelFunc
	mu       sync.Mutex
}

// NewLogModel creates a log viewer for one or more services.
func NewLogModel(services []ServiceLog) LogModel {
	return LogModel{
		Services:   services,
		Buffer:     []string{},
		Follow:     true,
		Timestamps: false,
		Wrap:       false,
		FilterText: "",
		Cursor:     0,
		width:      80,
	}
}

// Start begins streaming logs for all services.
func (m *LogModel) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancelFn = cancel

	for _, svc := range m.Services {
		go m.streamService(ctx, svc)
	}
}

// streamService reads log lines from a single service channel.
func (m *LogModel) streamService(ctx context.Context, svc ServiceLog) {
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-svc.Lines:
			if !ok {
				return
			}
			m.mu.Lock()
			m.Buffer = append(m.Buffer, line)
			if len(m.Buffer) > 10000 {
				m.Buffer = m.Buffer[len(m.Buffer)-10000:]
			}
			m.mu.Unlock()
		}
	}
}

// Stop cancels all log streams.
func (m *LogModel) Stop() {
	if m.cancelFn != nil {
		m.cancelFn()
	}
}

// Init satisfies tea.Model.
func (m *LogModel) Init() tea.Cmd { return nil }

// Update handles keyboard input for the log viewer.
func (m *LogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	}
	return m, nil
}

func (m *LogModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.Stop()
		return m, tea.Quit
	case "j", "down":
		if m.Cursor < len(m.Buffer)-1 {
			m.Cursor++
		}
	case "k", "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "g":
		m.Cursor = 0
	case "G":
		m.mu.Lock()
		m.Cursor = len(m.Buffer) - 1
		m.mu.Unlock()
	case "f":
		m.Follow = !m.Follow
	case "t":
		m.Timestamps = !m.Timestamps
	case "w":
		m.Wrap = !m.Wrap
	case "c":
		m.mu.Lock()
		m.Buffer = []string{}
		m.Cursor = 0
		m.mu.Unlock()
	case "/":
		// Filter mode - handled by external input capture
		// Placeholder: actual filter input handled via separate mechanism
	}
	return m, nil
}

// SetFilter updates the filter text.
func (m *LogModel) SetFilter(filter string) {
	m.FilterText = filter
}

// VisibleLines returns the filtered log lines suitable for display.
func (m *LogModel) VisibleLines() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lines []string
	for _, line := range m.Buffer {
		if m.FilterText == "" || strings.Contains(strings.ToLower(line), strings.ToLower(m.FilterText)) {
			lines = append(lines, line)
		}
	}
	return lines
}

// View renders the log viewer panel.
func (m *LogModel) View() string {
	var s []string
	s = append(s, lipgloss.NewStyle().Bold(true).Foreground(theme.ColorPrimary).Render("dockpose — logs"))
	s = append(s, "\n")
	s = append(s, m.renderStatusBar())
	s = append(s, "\n")
	s = append(s, theme.BorderStyle.Render(strings.Repeat("─", max(1, m.width-2))))
	s = append(s, "\n")
	s = append(s, m.renderLogLines())
	s = append(s, theme.HelpStyle.Render("j/k scroll | g/G top/bottom | f follow | t ts | w wrap | c clear | / filter | Esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

func (m *LogModel) renderStatusBar() string {
	var status []string
	status = append(status, theme.MutedStyle.Render("services: "))
	for i, svc := range m.Services {
		if i > 0 {
			status = append(status, theme.MutedStyle.Render(", "))
		}
		status = append(status, AccentStyle.Render(svc.ServiceName))
	}
	status = append(status, "  ")
	if m.Follow {
		status = append(status, theme.RunningStyle.Render("[follow]"))
	} else {
		status = append(status, theme.MutedStyle.Render("[follow]"))
	}
	status = append(status, "  ")
	if m.Timestamps {
		status = append(status, theme.RunningStyle.Render("[ts]"))
	} else {
		status = append(status, theme.MutedStyle.Render("[ts]"))
	}
	status = append(status, "  ")
	if m.Wrap {
		status = append(status, theme.RunningStyle.Render("[wrap]"))
	} else {
		status = append(status, theme.MutedStyle.Render("[wrap]"))
	}
	if m.FilterText != "" {
		status = append(status, "  ")
		status = append(status, WarningStyle.Render("filter: "+m.FilterText))
	}
	return lipgloss.JoinHorizontal(0, status...)
}

func (m *LogModel) renderLogLines() string {
	var s []string
	lines := m.VisibleLines()
	if len(lines) == 0 {
		return theme.MutedStyle.Render("(no logs)")
	}
	start := 0
	if len(lines) > m.Cursor {
		start = len(lines) - m.Cursor
	}
	visible := 50
	if start+visible > len(lines) {
		visible = len(lines) - start
	}
	for i := start; i < start+visible && i < len(lines); i++ {
		line := lines[i]
		if !m.Wrap {
			maxWidth := m.width - 10
			if maxWidth > 10 && len(line) > maxWidth {
				line = line[:maxWidth] + "…"
			}
		}
		prefix := "  "
		if i == len(lines)-1 {
			prefix = "▸ "
		}
		s = append(s, theme.NormalStyle.Render(prefix+line))
		s = append(s, "\n")
	}
	if len(lines) > visible {
		s = append(s, theme.MutedStyle.Render(fmt.Sprintf("(%d lines not shown)", len(lines)-visible)))
		s = append(s, "\n")
	}
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AccentStyle renders accent-colored text.
var AccentStyle = lipgloss.NewStyle().Foreground(theme.ColorAccent)

// WarningStyle renders warning-colored text.
var WarningStyle = lipgloss.NewStyle().Foreground(theme.ColorWarning)
