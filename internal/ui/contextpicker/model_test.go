package contextpicker

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEnterSelectsContextWithoutQuittingProgram(t *testing.T) {
	m := Model{
		Contexts: []Context{{Name: "default"}, {Name: "homelab", Endpoint: "ssh://host"}},
		Cursor:   1,
		Active:   true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		t.Fatalf("context picker should not emit tea.Quit; got %#v", cmd)
	}
	if m.Active {
		t.Fatal("expected picker to close")
	}
	if m.SelectedContext() != "homelab" {
		t.Fatalf("selected context = %q, want homelab", m.SelectedContext())
	}
}

func TestEnterWithNoContextsStaysOpen(t *testing.T) {
	m := Model{Active: true}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		t.Fatalf("unexpected command: %#v", cmd)
	}
	if !m.Active {
		t.Fatal("expected picker to stay open when there are no contexts")
	}
	if m.Error == "" {
		t.Fatal("expected an error for empty context list")
	}
}
