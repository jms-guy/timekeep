//go:build linux

package sql

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Gets database directory path for Linux
func getDatabasePath() (string, error) {
	cmd := exec.Command("sh", "-c", "logname || whoami")
	if user, err := cmd.Output(); err == nil {
		username := strings.TrimSpace(string(user))
		if username != "root" {
			dbPath := filepath.Join("home", username, ".local", "share", "timekeep", "timekeep.db")
			log.Printf("SERVICE DEBUG: Database path: %s", dbPath)

			return dbPath, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(home, ".local", "share", "timekeep", "timekeep.db")
	log.Printf("SERVICE DEBUG: Database path: %s", dbPath)

	return dbPath, nil
}
