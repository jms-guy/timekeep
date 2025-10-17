package logs

import (
	"log"
	"os"
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

	logger, f, err := CreateLogger(logPath)
	if err != nil {
		return nil, err
	}

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
