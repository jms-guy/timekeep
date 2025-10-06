//go:build linux

package config

import (
	"os"
	"path/filepath"
)

func getConfigLocation() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".local", "config", "timekeep", "config.json")

	return path, nil
}
