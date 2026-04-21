package stack

import (
	"path/filepath"
	"testing"
)

func TestRegistryLoadAndCache(t *testing.T) {
	root, err := filepath.Abs("../../testdata/fixtures")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	reg, err := LoadFromPaths([]string{root}, 3)
	if err != nil {
		t.Fatalf("LoadFromPaths: %v", err)
	}
	if reg.Count() == 0 {
		t.Fatal("expected at least one stack discovered from fixtures")
	}

	cachePath := filepath.Join(t.TempDir(), "registry.toml")
	if err := reg.CacheTo(cachePath); err != nil {
		t.Fatalf("CacheTo: %v", err)
	}

	loaded, err := LoadCache(cachePath)
	if err != nil {
		t.Fatalf("LoadCache: %v", err)
	}
	if loaded.Count() != reg.Count() {
		t.Fatalf("cache mismatch: got %d, want %d", loaded.Count(), reg.Count())
	}
}
