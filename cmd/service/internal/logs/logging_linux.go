//go:build linux

package logs

import (
	"log"
	"os"
	"path/filepath"
)

// Get path for logging file
func getLogPath() (string, error) {
	logDir := "/var/log/timekeep"
	return filepath.Join(logDir, "timekeep.log"), nil
}

func CreateLogger(logPath string) (*log.Logger, *os.File, error) {
	logger := log.Default()
	return logger, nil, nil
}
