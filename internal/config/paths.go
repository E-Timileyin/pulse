package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetPulseDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "pulse")
	case "darwin": //MacOS
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "pulse")
	default: //Linux
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, _ := os.UserHomeDir()
			configHome = filepath.Join(home, ".config")
		}
		return filepath.Join(configHome, "pulse")
	}
}

// Returns directory for state files
func GetStateDir() string {
	return filepath.Join(GetPulseDir(), "state")
}

// Returns directory for logs
func GetLogsDir() string {
	return filepath.Join(GetPulseDir(), "logs")
}

// EnsureDirs creates all required directories
func EnsureDirs() error {
	dirs := []string{GetPulseDir(), GetStateDir(), GetLogsDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
