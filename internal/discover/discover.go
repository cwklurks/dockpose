// Package discover finds Docker Compose stacks on disk.
package discover

import (
	"os"
	"path/filepath"
	"strings"

	composetypes "github.com/compose-spec/compose-go/v2/types"
	"gopkg.in/yaml.v3"
)

// StackCandidate is a discovered compose project path.
type StackCandidate struct {
	Name string
	Path string
}

// ParsedStack pairs a discovered candidate with a normalized compose project.
type ParsedStack struct {
	Candidate StackCandidate
	Project   *composetypes.Project
}

var skipDirs = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
}

// Discover walks each path up to scanDepth finding compose.yml/compose.yaml files.
// Stacks whose depends_on graph is not a valid DAG are skipped.
func Discover(scanPaths []string, scanDepth int) ([]StackCandidate, error) {
	var out []StackCandidate
	seen := map[string]bool{}
	for _, root := range scanPaths {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			absRoot = root
		}
		err = walk(absRoot, absRoot, 0, scanDepth, func(file string) {
			if seen[file] {
				return
			}
			seen[file] = true
			if !isValidDAG(file) {
				return
			}
			dir := filepath.Dir(file)
			out = append(out, StackCandidate{
				Name: filepath.Base(dir),
				Path: file,
			})
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func walk(root, dir string, depth, maxDepth int, onFile func(string)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		name := e.Name()
		full := filepath.Join(dir, name)
		if e.IsDir() {
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				continue
			}
			if depth >= maxDepth {
				continue
			}
			if err := walk(root, full, depth+1, maxDepth, onFile); err != nil {
				return err
			}
			continue
		}
		if name == "compose.yml" || name == "compose.yaml" {
			onFile(full)
		}
	}
	return nil
}

type composeDoc struct {
	Services map[string]struct {
		DependsOn any `yaml:"depends_on"`
	} `yaml:"services"`
}

func isValidDAG(file string) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	var doc composeDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	graph := map[string][]string{}
	for name, svc := range doc.Services {
		graph[name] = depsList(svc.DependsOn)
	}
	// cycle detection via DFS
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	for n := range graph {
		color[n] = white
	}
	var visit func(string) bool
	visit = func(n string) bool {
		color[n] = gray
		for _, m := range graph[n] {
			if _, ok := graph[m]; !ok {
				continue
			}
			if color[m] == gray {
				return false
			}
			if color[m] == white && !visit(m) {
				return false
			}
		}
		color[n] = black
		return true
	}
	for n := range graph {
		if color[n] == white {
			if !visit(n) {
				return false
			}
		}
	}
	return true
}

func depsList(v any) []string {
	switch t := v.(type) {
	case []any:
		var out []string
		for _, it := range t {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case map[string]any:
		var out []string
		for k := range t {
			out = append(out, k)
		}
		return out
	}
	return nil
}
