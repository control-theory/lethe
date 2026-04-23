package config

import (
	"os"
	"path/filepath"
)

const appName = "lethe"

// DataDir returns the platform-specific data directory for lethe,
// creating it if it does not exist.
func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// DBPath returns the path to the SQLite database file.
func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "memory.db"), nil
}

// LockPath returns the path to the singleton lockfile.
func LockPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "singleton.lock"), nil
}
