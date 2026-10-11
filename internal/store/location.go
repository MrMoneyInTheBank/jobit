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
	var base string
	switch platform {
	case "darwin":
		h, err := home()
		if err != nil {
			return "", err
		}
		base = filepath.Join(h, "Library", "Application Support")
	case "linux":
		base = getenv("XDG_DATA_HOME")
		if base == "" || !filepath.IsAbs(base) {
			h, err := home()
			if err != nil {
				return "", err
			}
			base = filepath.Join(h, ".local", "share")
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", platform)
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("application data directory must be absolute")
	}
	return filepath.Join(base, "jobit", "jobit.db"), nil
}

// DatabaseDSN reports the effective location without opening or creating it.
func DatabaseDSN(explicit *string) (string, error) {
	if explicit != nil && *explicit != "" {
		return *explicit, nil
	}
	if env := os.Getenv("JOBIT_DB_DSN"); env != "" {
		return env, nil
	}
	return DefaultDatabasePath()
}
