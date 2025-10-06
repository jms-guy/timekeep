//go:build linux

package sql

import (
	"os"
	"path/filepath"
)

// Gets database directory path for Linux
func getDatabasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dbPath := filepath.Join(home, ".local", "share", "timekeep", "timekeep.db")

	return dbPath, nil
}
