package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Logs struct {
	Logger  *log.Logger // Logging object
	LogFile *os.File    // Reference to the output log file
}

// Creates logger object, and log file reference
func NewLogs() (*Logs, error) {
	logPath, err := getLogPath()
	if err != nil {
		return nil, err
	}

	// #nosec G301
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, fmt.Errorf("ERROR: failed to create log directory: %w", err)
	}

	// #nosec -- Log file not security issue
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}

	logger := log.New(f, "", log.LstdFlags)

	return &Logs{Logger: logger, LogFile: f}, nil
}

func NewTestLogs() *Logs {
	return &Logs{Logger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags)}
}

// Closes any open log files
func (l *Logs) FileCleanup() {
	if l.LogFile != nil {
		l.Logger.Println("INFO: Closing log file connection")
		l.LogFile.Close()
	}
}
