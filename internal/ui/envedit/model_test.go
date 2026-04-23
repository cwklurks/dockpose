package envedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateEditsAndSavesEnvValue(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(composePath, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env")
	original := "# keep this comment\n\nAPI_TOKEN=\"old\"\nexport PASSWORD='secret'\nPLAIN=value\n"
	if err := os.WriteFile(envPath, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}

	model := New("demo", composePath)
	if len(model.Entries) != 6 {
		t.Fatalf("expected blank/comment/key entries to be preserved, got %#v", model.Entries)
	}
	if model.Cursor != 2 {
		t.Fatalf("expected cursor on first editable entry, got %d", model.Cursor)
	}

	next, _ := model.Update(key(tea.KeyEnter))
	model = next.(Model)
	if !model.Editing || model.EditValue != "old" {
		t.Fatalf("expected editing old value, got editing=%v value=%q", model.Editing, model.EditValue)
	}

	next, _ = model.Update(key(tea.KeyCtrlU))
	model = next.(Model)
	next, _ = model.Update(runes("new token"))
	model = next.(Model)
	next, _ = model.Update(key(tea.KeyEnter))
	model = next.(Model)
	if model.Editing {
		t.Fatal("expected edit mode to close after enter")
	}
	if model.Entries[2].Value != "new token" {
		t.Fatalf("expected committed value, got %q", model.Entries[2].Value)
	}

	model.Save()
	if model.SaveError != "" {
		t.Fatalf("save error: %s", model.SaveError)
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{
		"# keep this comment\n\n",
		"API_TOKEN=\"new token\"",
		"export PASSWORD='secret'",
		"PLAIN=value",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("saved .env missing %q:\n%s", want, got)
		}
	}
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("expected mode 0640 to be preserved, got %o", got)
	}
}

func TestSensitiveValuesAreMaskedUntilRevealed(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yaml")
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("API_KEY=abc123\nNORMAL=visible\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	model := New("demo", composePath)
	out := model.View()
	if strings.Contains(out, "abc123") {
		t.Fatalf("sensitive value leaked before reveal:\n%s", out)
	}

	next, _ := model.Update(runes("r"))
	model = next.(Model)
	out = model.View()
	if !strings.Contains(out, "abc123") {
		t.Fatalf("sensitive value should be visible after reveal:\n%s", out)
	}
}

func key(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
