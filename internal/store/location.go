package store

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultDatabasePath resolves user data storage without creating anything.
func DefaultDatabasePath() (string, error) {
	return databasePath(runtime.GOOS, os.Getenv, os.UserHomeDir)
}

func databasePath(platform string, getenv func(string) string, home func() (string, error)) (string, error) {
	var baseDirectory string
	switch platform {
	case "darwin":
		homeDirectory, err := home()
		if err != nil {
			return "", err
		}
		baseDirectory = filepath.Join(homeDirectory, "Library", "Application Support")
	case "linux":
		baseDirectory = getenv("XDG_DATA_HOME")
		if baseDirectory == "" || !filepath.IsAbs(baseDirectory) {
			homeDirectory, err := home()
			if err != nil {
				return "", err
			}
			baseDirectory = filepath.Join(homeDirectory, ".local", "share")
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", platform)
	}
	if !filepath.IsAbs(baseDirectory) {
		return "", fmt.Errorf("application data directory must be absolute")
	}
	return filepath.Join(baseDirectory, "jobit", "jobit.db"), nil
}
