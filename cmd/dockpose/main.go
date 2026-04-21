package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cwklurks/dockpose/internal/demo"
	"github.com/cwklurks/dockpose/internal/discover"
	"github.com/cwklurks/dockpose/internal/docker"
	"github.com/cwklurks/dockpose/internal/stack"
	"github.com/cwklurks/dockpose/internal/ui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	demoMode := flag.Bool("demo", false, "run with synthetic stacks; no Docker daemon required")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dockpose - a keyboard-driven TUI for managing Docker Compose stacks\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [paths...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                       # discover stacks under .\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s ~/homelab ~/projects  # scan multiple roots\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --demo                # try the TUI without Docker\n", os.Args[0])
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("dockpose %s\n", version)
		return
	}

	if *demoMode {
		runDemo()
		return
	}

	runReal(flag.Args())
}

func runDemo() {
	src := demo.New()
	model := ui.NewAppModelWithSource(src, true)
	model.SetStacks(src.Stacks())
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dockpose: TUI error: %v\n", err)
		os.Exit(1)
	}
}

func runReal(paths []string) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	ctx := context.Background()
	dockerClient, err := docker.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dockpose: failed to connect to Docker: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: try `dockpose --demo` to explore the TUI without Docker.\n")
		os.Exit(1)
	}

	candidates, err := discover.Discover(paths, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dockpose: failed to discover stacks: %v\n", err)
		os.Exit(1)
	}

	stacks := make([]stack.Stack, 0, len(candidates))
	for _, cand := range candidates {
		st, err := stack.ParseCompose(cand.Path)
		if err != nil {
			continue
		}
		_ = ctx // initial status comes from the polling loop
		stacks = append(stacks, *st)
	}

	src := docker.NewClientSource(dockerClient)
	model := ui.NewAppModelWithSource(src, false)
	model.SetStacks(stacks)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dockpose: TUI error: %v\n", err)
		os.Exit(1)
	}
}
