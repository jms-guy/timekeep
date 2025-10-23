//go:build windows

package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Get path for logging file
func getLogPath() (string, error) {
	logDir := `C:\ProgramData\TimeKeep\logs`
	return filepath.Join(logDir, "timekeep.log"), nil
}

func CreateLogger(logPath string) (*log.Logger, *os.File, error) {
	// #nosec G301
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, nil, fmt.Errorf("ERROR: failed to create log directory: %w", err)
	}

	// #nosec -- Log file not security issue
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, nil, err
	}

	logger := log.New(f, "", log.LstdFlags)

	return logger, f, nil
}
