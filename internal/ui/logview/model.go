// Package logview provides a streaming log viewer bubble tea component.
package logview

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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

// LogEntry is one received log line plus the time dockpose observed it.
type LogEntry struct {
	Text       string
	ReceivedAt time.Time
}

// LogModel is the bubble Tea model for the streaming log viewer.
type LogModel struct {
	Services   []ServiceLog
	Buffer     []LogEntry
	Follow     bool
	Timestamps bool
	Wrap       bool
	FilterText string
	Filtering  bool
	Cursor     int
	width      int

	cancelFn context.CancelFunc
	mu       sync.Mutex
}

// NewLogModel creates a log viewer for one or more services.
func NewLogModel(services []ServiceLog) LogModel {
	return LogModel{
		Services:   services,
		Buffer:     []LogEntry{},
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
			m.Buffer = append(m.Buffer, LogEntry{Text: line, ReceivedAt: time.Now()})
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
	if m.Filtering {
		m.handleFilterKey(msg)
		return m, nil
	}
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.Stop()
		return m, tea.Quit
	case "j", "down":
		m.Follow = false
		if m.Cursor < m.maxScrollStart() {
			m.Cursor++
		}
	case "k", "up":
		m.Follow = false
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "g":
		m.Follow = false
		m.Cursor = 0
	case "G":
		m.Follow = true
		m.Cursor = 0
	case "f":
		m.Follow = !m.Follow
		if m.Follow {
			m.Cursor = 0
		}
	case "t":
		m.Timestamps = !m.Timestamps
	case "w":
		m.Wrap = !m.Wrap
	case "c":
		m.mu.Lock()
		m.Buffer = []LogEntry{}
		m.Cursor = 0
		m.mu.Unlock()
	case "/":
		m.Filtering = true
		m.FilterText = ""
		m.Cursor = 0
	}
	return m, nil
}

func (m *LogModel) handleFilterKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEnter:
		m.Filtering = false
	case tea.KeyEsc:
		m.Filtering = false
		m.FilterText = ""
	case tea.KeyBackspace:
		if len(m.FilterText) > 0 {
			runes := []rune(m.FilterText)
			m.FilterText = string(runes[:len(runes)-1])
		}
	case tea.KeyCtrlU:
		m.FilterText = ""
	case tea.KeySpace:
		m.FilterText += " "
	case tea.KeyRunes:
		m.FilterText += string(msg.Runes)
	}
	m.Cursor = 0
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
	for _, entry := range m.Buffer {
		if m.FilterText == "" || strings.Contains(strings.ToLower(entry.Text), strings.ToLower(m.FilterText)) {
			line := entry.Text
			if m.Timestamps {
				line = entry.ReceivedAt.Format("15:04:05") + " " + line
			}
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
		label := "filter: " + m.FilterText
		if m.Filtering {
			label += "_"
		}
		status = append(status, WarningStyle.Render(label))
	} else if m.Filtering {
		status = append(status, "  ")
		status = append(status, WarningStyle.Render("filter: _"))
	}
	return lipgloss.JoinHorizontal(0, status...)
}

func (m *LogModel) renderLogLines() string {
	var s []string
	lines := m.VisibleLines()
	if len(lines) == 0 {
		return theme.MutedStyle.Render("(no logs)")
	}
	visible := visibleLogLines
	start := m.visibleStart(len(lines), visible)
	end := start + visible
	if end > len(lines) {
		end = len(lines)
	}
	for i := start; i < end; i++ {
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
	if start > 0 {
		s = append(s, theme.MutedStyle.Render(fmt.Sprintf("(%d lines above)", start)))
		s = append(s, "\n")
	}
	if end < len(lines) {
		s = append(s, theme.MutedStyle.Render(fmt.Sprintf("(%d lines below)", len(lines)-end)))
		s = append(s, "\n")
	}
	return lipgloss.JoinVertical(lipgloss.Left, s...)
}

const visibleLogLines = 50

func (m *LogModel) visibleStart(total, visible int) int {
	if total <= visible {
		return 0
	}
	if m.Follow {
		return total - visible
	}
	maxStart := total - visible
	if m.Cursor < 0 {
		return 0
	}
	if m.Cursor > maxStart {
		return maxStart
	}
	return m.Cursor
}

func (m *LogModel) maxScrollStart() int {
	total := len(m.VisibleLines())
	if total <= visibleLogLines {
		return 0
	}
	return total - visibleLogLines
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
