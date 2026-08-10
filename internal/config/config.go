// Package config defines the runtime configuration for a Pulse execution.
package config

import (
	"os"
	"path/filepath"
)

// Config holds the resolved runtime configuration for a single Pulse run.
type Config struct {
	// TargetPath is the absolute, cleaned path to the project being analysed.
	TargetPath string

	// JSON controls whether output is machine-readable JSON.
	JSON bool
}

// New builds a Config from the provided target path and output mode.
// If targetPath is empty, the current working directory is used.
// Relative paths are resolved to absolute paths.
func New(targetPath string, json bool) (*Config, error) {
	resolved, err := resolvePath(targetPath)
	if err != nil {
		return nil, err
	}
	return &Config{
		TargetPath: resolved,
		JSON:       json,
	}, nil
}

// resolvePath converts a raw path string into a clean absolute path.
func resolvePath(raw string) (string, error) {
	if raw == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Clean(cwd), nil
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(cwd, raw)), nil
}
