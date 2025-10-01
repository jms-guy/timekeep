//go:build linux

package sql

import (
	"log"
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
	log.Printf("SERVICE DEBUG: Database path: %s", dbPath)
	return dbPath, nil
}
