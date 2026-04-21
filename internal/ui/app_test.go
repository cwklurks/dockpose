package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cwklurks/dockpose/internal/demo"
)

// TestDemoModelRenders is a smoke test: it builds a demo source, wires it
// into AppModel, and exercises View() across each top-level screen.
func TestDemoModelRenders(t *testing.T) {
	src := demo.New()
	m := NewAppModelWithSource(src, true)
	m.SetStacks(src.Stacks())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = mm.(AppModel)

	out := m.View()
	if !strings.Contains(out, "dockpose") {
		t.Fatalf("expected dockpose header, got:\n%s", out)
	}
	if !strings.Contains(out, "demo") {
		t.Fatalf("expected demo chip, got:\n%s", out)
	}
	if !strings.Contains(out, "STATUS") {
		t.Fatalf("expected STATUS column header, got:\n%s", out)
	}
	if !strings.Contains(out, "media") {
		t.Fatalf("expected fixture stack 'media', got:\n%s", out)
	}

	// help overlay
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = mm.(AppModel)
	out = m.View()
	if !strings.Contains(out, "Stack list") {
		t.Fatalf("expected help overlay, got:\n%s", out)
	}

	// close help, open detail
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(AppModel)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(AppModel)
	out = m.View()
	if !strings.Contains(out, "Dependency Graph") {
		t.Fatalf("expected detail view, got:\n%s", out)
	}
}

// TestDemoFrameDump renders one demo frame and prints it. Run with -v to see.
func TestDemoFrameDump(t *testing.T) {
	src := demo.New()
	for i := 0; i < 6; i++ {
		src.Tick()
	}
	m := NewAppModelWithSource(src, true)
	m.SetStacks(src.Stacks())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 110, Height: 40})
	m = mm.(AppModel)
	mm, _ = m.Update(tickMsg{})
	m = mm.(AppModel)
	t.Log("\n" + m.View())
}

func TestStackListFilterNarrows(t *testing.T) {
	src := demo.New()
	m := NewAppModelWithSource(src, true)
	m.SetStacks(src.Stacks())
	if got := len(m.visibleStacks()); got != len(src.Stacks()) {
		t.Fatalf("expected %d stacks, got %d", len(src.Stacks()), got)
	}
	m.filter = "media"
	got := m.visibleStacks()
	if len(got) != 1 || got[0].Name != "media" {
		t.Fatalf("filter 'media' should yield only the media stack, got %v", got)
	}
}
