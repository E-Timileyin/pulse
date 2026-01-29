package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPulseDir(t *testing.T) {
	dir := GetPulseDir()
	if dir == "" {
		t.Error("GetPulseDir returned empty string")
	}
	// Should contain "pulse" in path
	if !strings.Contains(strings.ToLower(dir), "pulse") {
		t.Errorf("Expected path to contain 'pulse', got: %s", dir)
	}
}

func TestGetStateDir(t *testing.T) {
	dir := GetStateDir()
	if dir == "" {
		t.Error("GetStateDir returned empty string")
	}
	if !strings.HasSuffix(dir, "state") {
		t.Errorf("Expected path to end with 'state', got: %s", dir)
	}
	// State dir should be under pulse dir
	pulseDir := GetPulseDir()
	if !strings.HasPrefix(dir, pulseDir) {
		t.Errorf("StateDir should be under SurgeDir. StateDir: %s, SurgeDir: %s", dir, pulseDir)
	}
}

func TestGetLogsDir(t *testing.T) {
	dir := GetLogsDir()
	if dir == "" {
		t.Error("GetLogsDir returned empty string")
	}
	if !strings.HasSuffix(dir, "logs") {
		t.Errorf("Expected path to end with 'logs', got: %s", dir)
	}
	// Logs dir should be under pulse dir
	pulseDir := GetPulseDir()
	if !strings.HasPrefix(dir, pulseDir) {
		t.Errorf("LogsDir should be under SurgeDir. LogsDir: %s, SurgeDir: %s", dir, pulseDir)
	}
}

func TestEnsureDirs(t *testing.T) {
	err := EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Verify all directories exist
	dirs := []string{GetPulseDir(), GetStateDir(), GetLogsDir()}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		} else if err != nil {
			t.Errorf("Error checking directory %s: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("Path exists but is not a directory: %s", dir)
		}
	}
}

// Extended tests for cross-platform path handling

func TestGetPulseDir_AbsolutePath(t *testing.T) {
	dir := GetPulseDir()
	if !filepath.IsAbs(dir) {
		t.Errorf("GetPulseDir should return absolute path, got: %s", dir)
	}
}

func TestGetStateDir_AbsolutePath(t *testing.T) {
	dir := GetStateDir()
	if !filepath.IsAbs(dir) {
		t.Errorf("GetStateDir should return absolute path, got: %s", dir)
	}
}

func TestGetLogsDir_AbsolutePath(t *testing.T) {
	dir := GetLogsDir()
	if !filepath.IsAbs(dir) {
		t.Errorf("GetLogsDir should return absolute path, got: %s", dir)
	}
}

func TestPathConsistency(t *testing.T) {
	// Multiple calls should return the same paths
	dir1 := GetPulseDir()
	dir2 := GetPulseDir()
	if dir1 != dir2 {
		t.Errorf("GetPulseDir should return consistent paths: %s vs %s", dir1, dir2)
	}

	state1 := GetStateDir()
	state2 := GetStateDir()
	if state1 != state2 {
		t.Errorf("GetStateDir should return consistent paths: %s vs %s", state1, state2)
	}

	logs1 := GetLogsDir()
	logs2 := GetLogsDir()
	if logs1 != logs2 {
		t.Errorf("GetLogsDir should return consistent paths: %s vs %s", logs1, logs2)
	}
}

func TestDirectoryHierarchy(t *testing.T) {
	pulseDir := GetPulseDir()
	stateDir := GetStateDir()
	logsDir := GetLogsDir()

	// State and logs should be subdirectories of pulse dir
	expectedStateDir := filepath.Join(pulseDir, "state")
	expectedLogsDir := filepath.Join(pulseDir, "logs")

	if stateDir != expectedStateDir {
		t.Errorf("StateDir should be %s, got: %s", expectedStateDir, stateDir)
	}

	if logsDir != expectedLogsDir {
		t.Errorf("LogsDir should be %s, got: %s", expectedLogsDir, logsDir)
	}
}

func TestEnsureDirs_Idempotent(t *testing.T) {
	// EnsureDirs should be safe to call multiple times
	for i := 0; i < 3; i++ {
		err := EnsureDirs()
		if err != nil {
			t.Errorf("EnsureDirs failed on call %d: %v", i+1, err)
		}
	}
}

func TestPathsNoTrailingSlash(t *testing.T) {
	dirs := []struct {
		name string
		path string
	}{
		{"SurgeDir", GetPulseDir()},
		{"StateDir", GetStateDir()},
		{"LogsDir", GetLogsDir()},
	}

	for _, d := range dirs {
		if strings.HasSuffix(d.path, string(filepath.Separator)) {
			t.Errorf("%s should not have trailing separator: %s", d.name, d.path)
		}
	}
}
