// Package dag renders dependency graphs for compose stacks as ASCII art.
package dag

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/cwklurks/dockpose/internal/stack"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// RenderDependencyGraph renders a layered ASCII-art DAG of the given services.
// status maps service name to status keyword (running/stopped/unhealthy/...).
func RenderDependencyGraph(services map[string]stack.ServiceConfig, status map[string]string) string {
	if len(services) == 0 {
		return theme.MutedStyle.Render("(no services)")
	}

	adj, err := stack.BuildGraph(services)
	if err != nil {
		return theme.StoppedStyle.Render("dependency cycle: " + err.Error())
	}

	layers := topologicalLayers(services, adj)
	if len(layers) == 0 {
		return theme.MutedStyle.Render("(no services)")
	}

	nodeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(0, 1)

	var cols []string
	for _, layer := range layers {
		sort.Strings(layer)
		var rendered []string
		for _, name := range layer {
			ind := statusIndicator(status[name])
			label := ind + " " + name
			rendered = append(rendered, nodeStyle.Render(label))
		}
		col := lipgloss.JoinVertical(lipgloss.Left, rendered...)
		cols = append(cols, col)
	}

	// Interleave columns with arrow connectors.
	arrow := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(" → ")
	var parts []string
	for i, c := range cols {
		parts = append(parts, c)
		if i < len(cols)-1 {
			parts = append(parts, arrow)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

// statusIndicator returns a glyph based on status keyword.
func statusIndicator(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch {
	case strings.Contains(s, "unhealthy"), strings.Contains(s, "starting"), strings.Contains(s, "restarting"):
		return theme.UnhealthyStyle.Render("◐")
	case strings.Contains(s, "running"), strings.Contains(s, "healthy"), strings.Contains(s, "up"):
		return theme.RunningStyle.Render("●")
	default:
		return theme.StoppedStyle.Render("○")
	}
}

// topologicalLayers groups nodes into layers where each layer contains nodes
// whose dependencies are all in earlier layers.
func topologicalLayers(services map[string]stack.ServiceConfig, adj map[string][]string) [][]string {
	// In-degree per node (number of deps that point to it).
	inDeg := make(map[string]int, len(adj))
	for n := range adj {
		inDeg[n] = 0
	}
	for _, children := range adj {
		for _, c := range children {
			inDeg[c]++
		}
	}

	remaining := make(map[string]int, len(inDeg))
	for k, v := range inDeg {
		remaining[k] = v
	}

	var layers [][]string
	for len(remaining) > 0 {
		var layer []string
		for n, d := range remaining {
			if d == 0 {
				layer = append(layer, n)
			}
		}
		if len(layer) == 0 {
			// Cycle guard — shouldn't happen after BuildGraph.
			break
		}
		sort.Strings(layer)
		for _, n := range layer {
			delete(remaining, n)
			for _, c := range adj[n] {
				if _, ok := remaining[c]; ok {
					remaining[c]--
				}
			}
		}
		layers = append(layers, layer)
	}
	return layers
}
