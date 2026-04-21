package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.ScanDepth != 3 {
		t.Errorf("ScanDepth: want 3, got %d", c.ScanDepth)
	}
	if c.PollInterval != 2*time.Second {
		t.Errorf("PollInterval: want 2s, got %s", c.PollInterval)
	}
	if c.Theme != "default" {
		t.Errorf("Theme: want default, got %q", c.Theme)
	}
	if len(c.ScanPaths) == 0 {
		t.Error("ScanPaths should not be empty")
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(cfg, Default()) {
		t.Errorf("missing file should yield Default(); got %+v", cfg)
	}
}

func TestRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	want := Config{
		ScanPaths:    []string{"/srv/stacks", "/home/a/docker"},
		ScanDepth:    5,
		PollInterval: 500 * time.Millisecond,
		LogBuffer:    20_000,
		Theme:        "dracula",
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("roundtrip mismatch:\nwant %+v\n got %+v", want, got)
	}
}

func TestSaveCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deeper", "config.toml")
	if err := Save(path, Default()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestDefaultPathUsesXDG(t *testing.T) {
	t.Setenv("DOCKPOSE_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test-dp")
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	want := "/tmp/xdg-test-dp/dockpose/config.toml"
	if got != want {
		t.Errorf("DefaultPath: want %q got %q", want, got)
	}
}

func TestDefaultPathOverride(t *testing.T) {
	t.Setenv("DOCKPOSE_CONFIG_DIR", "/tmp/override")
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if got != "/tmp/override/config.toml" {
		t.Errorf("override: got %q", got)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	cases := map[string]string{
		"":          "",
		"/abs/path": "/abs/path",
		"~":         home,
		"~/docker":  filepath.Join(home, "docker"),
	}
	for in, want := range cases {
		if got := ExpandPath(in); got != want {
			t.Errorf("ExpandPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolvedScanPaths(t *testing.T) {
	home, _ := os.UserHomeDir()
	c := Config{ScanPaths: []string{"~/docker", "/abs"}}
	got := c.ResolvedScanPaths()
	want := []string{filepath.Join(home, "docker"), "/abs"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ResolvedScanPaths: got %v want %v", got, want)
	}
}

func TestLoadRejectsBadToml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("= not valid toml"), 0o644); err != nil {
		t.Fatalf("prep: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error on bad toml")
	}
}
