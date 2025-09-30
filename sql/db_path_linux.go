//go:build linux

package sql

import (
	"os"
	"path/filepath"
)

// Gets database directory path for Windows
func getDatabasePath() (string, error) {
	// User-specific database (recommended)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "timekeep", "timekeep.db"), nil
}
