// dockpose-record drives the demo TUI through a scripted scenario and
// writes the result as an asciinema v2 cast file.
//
// Usage:
//
//	go run ./cmd/dockpose-record > docs/media/demo.cast
//
// Play with:
//
//	asciinema play docs/media/demo.cast
//
// Convert to GIF (requires `agg`):
//
//	agg docs/media/demo.cast docs/media/demo.gif
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/cwklurks/dockpose/internal/demo"
	"github.com/cwklurks/dockpose/internal/ui"
)

const (
	width  = 120
	height = 26
)

// step describes one scripted action and the dwell time before the next.
type step struct {
	dwell time.Duration
	apply func(m ui.AppModel) ui.AppModel
}

func main() {
	// Force truecolor regardless of TTY detection so the cast captures
	// lipgloss styling rather than stripped plaintext.
	lipgloss.SetColorProfile(termenv.TrueColor)

	src := demo.New()
	m := ui.NewAppModelWithSource(src, true)
	m.SetStacks(src.Stacks())
	m = applyMsg(m, tea.WindowSizeMsg{Width: width, Height: height})
	ctx := context.Background()
	m.Refresh(ctx)

	steps := []step{
		// Initial frame.
		{dwell: 1500 * time.Millisecond, apply: tickStep(src)},
		// Walk down the list.
		{dwell: 700 * time.Millisecond, apply: keyStep("j")},
		{dwell: 700 * time.Millisecond, apply: keyStep("j")},
		{dwell: 700 * time.Millisecond, apply: keyStep("j")},
		{dwell: 1200 * time.Millisecond, apply: tickStep(src)},
		// Open detail.
		{dwell: 1800 * time.Millisecond, apply: keyStep("enter")},
		// Move through services.
		{dwell: 700 * time.Millisecond, apply: keyStep("j")},
		{dwell: 700 * time.Millisecond, apply: keyStep("j")},
		{dwell: 1200 * time.Millisecond, apply: tickStep(src)},
		// Inspect a container.
		{dwell: 1800 * time.Millisecond, apply: keyStep("i")},
		{dwell: 800 * time.Millisecond, apply: keyStep("esc")},
		// Back to list, open help.
		{dwell: 600 * time.Millisecond, apply: keyStep("esc")},
		{dwell: 1800 * time.Millisecond, apply: keyStep("?")},
		{dwell: 800 * time.Millisecond, apply: keyStep("esc")},
		// Filter for "media".
		{dwell: 600 * time.Millisecond, apply: keyStep("/")},
		{dwell: 250 * time.Millisecond, apply: typeRune('m')},
		{dwell: 250 * time.Millisecond, apply: typeRune('e')},
		{dwell: 250 * time.Millisecond, apply: typeRune('d')},
		{dwell: 250 * time.Millisecond, apply: typeRune('i')},
		{dwell: 250 * time.Millisecond, apply: typeRune('a')},
		{dwell: 1200 * time.Millisecond, apply: keyStep("enter")},
		{dwell: 800 * time.Millisecond, apply: keyStep("esc")},
		// Try a destructive action (gets demo toast).
		{dwell: 1500 * time.Millisecond, apply: keyStep("d")},
		{dwell: 1500 * time.Millisecond, apply: tickStep(src)},
		{dwell: 1200 * time.Millisecond, apply: tickStep(src)},
	}

	enc := newCastEncoder(os.Stdout, width, height)
	enc.writeHeader()
	enc.writeFrame(0, m.View())

	t := 0.0
	for _, s := range steps {
		m = s.apply(m)
		t += s.dwell.Seconds()
		enc.writeFrame(t, m.View())
	}
	// Hold the final frame for a beat.
	t += 1.5
	enc.writeFrame(t, m.View())
}

func applyMsg(m ui.AppModel, msg tea.Msg) ui.AppModel {
	mm, _ := m.Update(msg)
	switch v := mm.(type) {
	case ui.AppModel:
		return v
	case *ui.AppModel:
		return *v
	default:
		return m
	}
}

func keyStep(s string) func(ui.AppModel) ui.AppModel {
	return func(m ui.AppModel) ui.AppModel {
		var msg tea.KeyMsg
		switch s {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "/":
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		case "?":
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
		}
		return applyMsg(m, msg)
	}
}

func typeRune(r rune) func(ui.AppModel) ui.AppModel {
	return func(m ui.AppModel) ui.AppModel {
		return applyMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func tickStep(_ *demo.Source) func(ui.AppModel) ui.AppModel {
	return func(m ui.AppModel) ui.AppModel {
		m.Refresh(context.Background())
		return m
	}
}

// castEncoder writes asciinema v2 cast events to an io.Writer.
type castEncoder struct {
	w        *os.File
	previous string
}

func newCastEncoder(w *os.File, width, height int) *castEncoder {
	return &castEncoder{w: w}
}

func (e *castEncoder) writeHeader() {
	hdr := map[string]any{
		"version":   2,
		"width":     width,
		"height":    height,
		"timestamp": time.Now().Unix(),
		"env":       map[string]string{"SHELL": "/bin/bash", "TERM": "xterm-256color"},
		"title":     "dockpose --demo",
	}
	b, _ := json.Marshal(hdr)
	if _, err := fmt.Fprintln(e.w, string(b)); err != nil {
		panic(err)
	}
}

// writeFrame emits a clear-screen + frame at time t.
func (e *castEncoder) writeFrame(t float64, frame string) {
	const clear = "\x1b[2J\x1b[H"
	payload := clear + frame
	b, _ := json.Marshal(payload)
	// asciinema event: [time, "o", text]
	if _, err := fmt.Fprintf(e.w, "[%.3f, \"o\", %s]\n", t, string(b)); err != nil {
		panic(err)
	}
}
