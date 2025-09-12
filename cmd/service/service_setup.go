package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jms-guy/timekeep/internal/repository"
	mysql "github.com/jms-guy/timekeep/sql"
)

// Service context
type timekeepService struct {
	prRepo         repository.ProgramRepository // Repository for tracked_programs database queries
	asRepo         repository.ActiveRepository  //Repository for active_sessions database queries
	hsRepo         repository.HistoryRepository // Repository for session_history database queries
	logger         *log.Logger                  // Logging object
	logFile        *os.File                     // Reference to the output log file
	psProcess      *exec.Cmd                    // The running WMI powershell script
	activeSessions map[string]map[string]bool   // Map of active sessions & their PIDs
	shutdown       chan struct{}                // Shutdown channel
}

func ServiceSetup() (*timekeepService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	logger, f, err := NewLogger()
	if err != nil {
		return nil, err
	}

	service := NewTimekeepService(store, store, store, logger, f)

	return service, nil
}

// Creates new service instance
func NewTimekeepService(pr repository.ProgramRepository, ar repository.ActiveRepository, hr repository.HistoryRepository, logger *log.Logger, f *os.File) *timekeepService {
	return &timekeepService{
		prRepo:         pr,
		asRepo:         ar,
		hsRepo:         hr,
		logger:         logger,
		logFile:        f,
		psProcess:      nil,
		activeSessions: make(map[string]map[string]bool),
		shutdown:       make(chan struct{}),
	}
}

// Creates logger object, and log file reference
func NewLogger() (*log.Logger, *os.File, error) {
	logPath, err := getLogPath()
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, nil, fmt.Errorf("ERROR: failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, err
	}

	logger := log.New(f, "", log.LstdFlags)

	return logger, f, nil
}

func getLogPath() (string, error) {
	logDir := `C:\ProgramData\TimeKeep\logs`
	return filepath.Join(logDir, "timekeep.log"), nil
}
