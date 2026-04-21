package ui

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cwklurks/dockpose/internal/discover"
	"github.com/cwklurks/dockpose/internal/docker"
	"github.com/cwklurks/dockpose/internal/stack"
)

// TestRealSourceWiring exercises the production code path end to end:
// real Docker client -> ClientSource -> AppModel -> first render. Skips
// when Docker isn't reachable.
func TestRealSourceWiring(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker binary not on PATH")
	}
	cli, err := docker.New()
	if err != nil {
		t.Skipf("docker.New: %v", err)
	}
	if _, err := cli.Ping(context.Background()); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	candidates, err := discover.Discover([]string{"/tmp/dpsmoke"}, 3)
	if err != nil {
		t.Skipf("discover: %v", err)
	}
	if len(candidates) == 0 {
		t.Skip("no fixture stack at /tmp/dpsmoke")
	}
	stacks := make([]stack.Stack, 0, len(candidates))
	for _, c := range candidates {
		st, err := stack.ParseCompose(c.Path)
		if err != nil {
			t.Fatalf("parse %s: %v", c.Path, err)
		}
		stacks = append(stacks, *st)
	}

	src := docker.NewClientSource(cli)
	m := NewAppModelWithSource(src, false)
	m.SetStacks(stacks)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = mm.(AppModel)
	m.Refresh(context.Background())

	out := m.View()
	if !strings.Contains(out, "dockpose") {
		t.Fatalf("expected dockpose header, got:\n%s", out)
	}
	if strings.Contains(out, "demo") {
		t.Fatalf("real mode should not show demo chip, got:\n%s", out)
	}
	if !strings.Contains(out, "STATUS") {
		t.Fatalf("expected STATUS column, got:\n%s", out)
	}
}
