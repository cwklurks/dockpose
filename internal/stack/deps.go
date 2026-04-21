package stack

import (
	"fmt"
	"sort"
)

// ServiceConfig is the minimal service description used for dependency graph building.
type ServiceConfig struct {
	DependsOn []string
}

// BuildGraph returns an adjacency list mapping dependency -> dependents.
// If a cycle is detected, it returns an error naming the services involved.
func BuildGraph(services map[string]ServiceConfig) (map[string][]string, error) {
	adj := make(map[string][]string, len(services))
	for name := range services {
		adj[name] = nil
	}
	for name, cfg := range services {
		for _, dep := range cfg.DependsOn {
			if _, ok := services[dep]; !ok {
				continue
			}
			adj[dep] = append(adj[dep], name)
		}
	}
	for _, children := range adj {
		sort.Strings(children)
	}
	if cycle := findCycle(adj); len(cycle) > 0 {
		return nil, fmt.Errorf("cycle detected: %v", cycle)
	}
	return adj, nil
}

// TopologicalOrder returns a deterministic start order using Kahn's algorithm.
func TopologicalOrder(adj map[string][]string) ([]string, error) {
	inDeg := map[string]int{}
	for n := range adj {
		inDeg[n] = 0
	}
	for _, children := range adj {
		for _, c := range children {
			inDeg[c]++
		}
	}
	var queue []string
	for n, d := range inDeg {
		if d == 0 {
			queue = append(queue, n)
		}
	}
	sort.Strings(queue)

	var out []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		out = append(out, n)
		next := []string{}
		for _, c := range adj[n] {
			inDeg[c]--
			if inDeg[c] == 0 {
				next = append(next, c)
			}
		}
		sort.Strings(next)
		queue = append(queue, next...)
		sort.Strings(queue)
	}
	if len(out) != len(adj) {
		return nil, fmt.Errorf("cycle in graph")
	}
	return out, nil
}

// DetectCycles returns true if the directed graph contains a cycle.
func DetectCycles(adj map[string][]string) bool {
	return len(findCycle(adj)) > 0
}

func findCycle(adj map[string][]string) []string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	for n := range adj {
		color[n] = white
	}
	var stack []string
	var cycle []string
	var visit func(string) bool
	visit = func(n string) bool {
		color[n] = gray
		stack = append(stack, n)
		for _, m := range adj[n] {
			if _, ok := color[m]; !ok {
				continue
			}
			if color[m] == gray {
				for i, s := range stack {
					if s == m {
						cycle = append([]string{}, stack[i:]...)
						break
					}
				}
				return true
			}
			if color[m] == white && visit(m) {
				return true
			}
		}
		stack = stack[:len(stack)-1]
		color[n] = black
		return false
	}

	names := make([]string, 0, len(adj))
	for n := range adj {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if color[n] == white {
			if visit(n) {
				return cycle
			}
		}
	}
	return nil
}
