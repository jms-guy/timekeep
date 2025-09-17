//go:build windows

package logs

import "path/filepath"

// Get path for logging file
func getLogPath() (string, error) {
	logDir := `C:\ProgramData\TimeKeep\logs`
	return filepath.Join(logDir, "timekeep.log"), nil
}
