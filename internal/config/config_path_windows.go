//go:build windows

package config

import "path/filepath"

func getConfigLocation() (string, error) {
	configDir := `C:\ProgramData\Timekeep\config`
	return filepath.Join(configDir, "config.json"), nil
}
