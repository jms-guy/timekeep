//go:build linux

package logs

import "path/filepath"

// Get path for logging file
func getLogPath() (string, error) {
	logDir := "/var/log/timekeep/logs"
	return filepath.Join(logDir, "timekeep.log"), nil
}
