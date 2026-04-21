// Package config loads and persists dockpose user configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// ConfigFileName is the canonical config file name on disk.
const ConfigFileName = "config.toml"

// AppName is the config sub-directory name (under $XDG_CONFIG_HOME).
const AppName = "dockpose"

// Config is the top-level user configuration model.
type Config struct {
	ScanPaths    []string      `toml:"scan_paths"`
	ScanDepth    int           `toml:"scan_depth"`
	PollInterval time.Duration `toml:"poll_interval"`
	LogBuffer    int           `toml:"log_buffer"`
	Theme        string        `toml:"theme"`
}

// Default returns the baseline configuration used when no config file is present.
func Default() Config {
	return Config{
		ScanPaths:    []string{"~/docker", "~/homelab", "~/projects", "~/stacks"},
		ScanDepth:    3,
		PollInterval: 2 * time.Second,
		LogBuffer:    10_000,
		Theme:        "default",
	}
}

// Dir returns the XDG-compliant dockpose config directory.
func Dir() (string, error) {
	if dir := os.Getenv("DOCKPOSE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, AppName), nil
}

// DefaultPath returns the default config file path under $XDG_CONFIG_HOME/dockpose.
func DefaultPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// ExpandPath resolves a leading ~ (and ~/user/...) to an absolute path.
// An empty string is returned unchanged.
func ExpandPath(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	if len(p) > 1 && (p[1] == '/' || p[1] == filepath.Separator) {
		return filepath.Join(home, p[2:])
	}
	return p
}

// Load decodes a TOML config file, returning defaults when the file is absent.
// An empty path uses DefaultPath().
func Load(path string) (Config, error) {
	if path == "" {
		dp, err := DefaultPath()
		if err != nil {
			return Config{}, err
		}
		path = dp
	}
	cfg := Default()
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("stat config: %w", err)
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if cfg.ScanDepth <= 0 {
		cfg.ScanDepth = Default().ScanDepth
	}
	return cfg, nil
}

// Save encodes cfg as TOML and writes atomically to path. Parent directory is
// created if missing. An empty path uses DefaultPath().
func Save(path string, cfg Config) error {
	if path == "" {
		dp, err := DefaultPath()
		if err != nil {
			return err
		}
		path = dp
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.toml")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := toml.NewEncoder(tmp).Encode(cfg); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("encode config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// ResolvedScanPaths returns ScanPaths with ~ expanded. Existence is not checked.
func (c Config) ResolvedScanPaths() []string {
	out := make([]string, 0, len(c.ScanPaths))
	for _, p := range c.ScanPaths {
		out = append(out, ExpandPath(p))
	}
	return out
}
