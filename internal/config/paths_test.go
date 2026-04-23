package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataDir(t *testing.T) {
	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("lethe")) {
		t.Errorf("DataDir() = %q, want suffix %q", dir, "lethe")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory, got file")
	}
}

func TestDBPath(t *testing.T) {
	p, err := DBPath()
	if err != nil {
		t.Fatalf("DBPath() error: %v", err)
	}
	if !strings.HasSuffix(p, filepath.Join("lethe", "memory.db")) {
		t.Errorf("DBPath() = %q, want suffix %q", p, "lethe/memory.db")
	}
}

func TestLockPath(t *testing.T) {
	p, err := LockPath()
	if err != nil {
		t.Fatalf("LockPath() error: %v", err)
	}
	if !strings.HasSuffix(p, filepath.Join("lethe", "singleton.lock")) {
		t.Errorf("LockPath() = %q, want suffix %q", p, "lethe/singleton.lock")
	}
}
