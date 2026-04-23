package logview

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFollowRendersNewestLines(t *testing.T) {
	m := NewLogModel(nil)
	m.width = 120
	for i := 0; i < 60; i++ {
		m.Buffer = append(m.Buffer, LogEntry{
			Text:       "line " + twoDigits(i),
			ReceivedAt: time.Date(2026, 4, 23, 12, 0, i%60, 0, time.UTC),
		})
	}

	out := m.renderLogLines()
	if strings.Contains(out, "line 00") {
		t.Fatalf("follow view should show tail, not oldest lines:\n%s", out)
	}
	if !strings.Contains(out, "line 59") {
		t.Fatalf("follow view should include newest line:\n%s", out)
	}
	if !strings.Contains(out, "(10 lines above)") {
		t.Fatalf("expected count of hidden older lines:\n%s", out)
	}
}

func TestFilterInputNarrowsVisibleLines(t *testing.T) {
	m := NewLogModel(nil)
	m.Buffer = []LogEntry{
		{Text: "api ready", ReceivedAt: time.Now()},
		{Text: "worker ready", ReceivedAt: time.Now()},
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.Filtering {
		t.Fatal("expected slash to enter filter mode")
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("api")})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	lines := m.VisibleLines()
	if len(lines) != 1 || lines[0] != "api ready" {
		t.Fatalf("expected only api line, got %#v", lines)
	}
}

func TestTimestampToggleFormatsVisibleLines(t *testing.T) {
	m := NewLogModel(nil)
	m.Buffer = []LogEntry{{
		Text:       "api ready",
		ReceivedAt: time.Date(2026, 4, 23, 9, 8, 7, 0, time.UTC),
	}}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	lines := m.VisibleLines()
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "09:08:07 api ready") {
		t.Fatalf("expected timestamped line, got %#v", lines)
	}
}

func twoDigits(i int) string {
	return fmt.Sprintf("%02d", i)
}
