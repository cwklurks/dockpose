package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/cwklurks/dockpose/internal/discover"
)

// Registry stores the active in-memory stack list.
type Registry struct {
	Stacks []Stack `toml:"stacks"`
}

// Count returns the number of known stacks.
func (r Registry) Count() int {
	return len(r.Stacks)
}

// LoadFromPaths discovers compose files under scanPaths and parses each into a Stack.
// Candidates that fail to parse are skipped.
func LoadFromPaths(scanPaths []string, scanDepth int) (*Registry, error) {
	candidates, err := discover.Discover(scanPaths, scanDepth)
	if err != nil {
		return nil, fmt.Errorf("discover: %w", err)
	}
	reg := &Registry{}
	for _, c := range candidates {
		s, err := ParseCompose(c.Path)
		if err != nil {
			continue
		}
		reg.Stacks = append(reg.Stacks, *s)
	}
	return reg, nil
}

// CacheTo writes the registry to cachedPath as TOML (atomic via rename).
func (r *Registry) CacheTo(cachedPath string) error {
	if err := os.MkdirAll(filepath.Dir(cachedPath), 0o755); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(cachedPath), ".registry-*.toml")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := toml.NewEncoder(tmp).Encode(r); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("encode registry: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, cachedPath); err != nil {
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// LoadCache reads a previously cached registry from cachedPath.
func LoadCache(cachedPath string) (*Registry, error) {
	reg := &Registry{}
	if _, err := toml.DecodeFile(cachedPath, reg); err != nil {
		return nil, fmt.Errorf("decode registry: %w", err)
	}
	return reg, nil
}
