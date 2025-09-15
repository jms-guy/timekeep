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
	asRepo         repository.ActiveRepository  // Repository for active_sessions database queries
	hsRepo         repository.HistoryRepository // Repository for session_history database queries
	logger         *log.Logger                  // Logging object
	logFile        *os.File                     // Reference to the output log file
	psProcess      *exec.Cmd                    // The running OS-specific script for process monitoring
	activeSessions map[string]map[string]bool   // Map of active sessions & their PIDs
	shutdown       chan struct{}                // Shutdown channel
	daemon         DaemonManager                // Embedded daemon.Daemon struct wrapped by interface
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

	d, err := NewDaemonManager()
	if err != nil {
		return nil, err
	}

	service := NewTimekeepService(store, store, store, logger, f, d)

	return service, nil
}

func TestServiceSetup() (*timekeepService, error) {
	db, err := mysql.OpenTestDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	logger := log.New(os.Stdout, "DEBUG: ", log.LstdFlags)

	service := NewTimekeepService(store, store, store, logger, nil, nil)

	return service, nil
}

// Creates new service instance
func NewTimekeepService(pr repository.ProgramRepository, ar repository.ActiveRepository, hr repository.HistoryRepository, logger *log.Logger, f *os.File, d DaemonManager) *timekeepService {
	return &timekeepService{
		prRepo:         pr,
		asRepo:         ar,
		hsRepo:         hr,
		logger:         logger,
		logFile:        f,
		psProcess:      nil,
		activeSessions: make(map[string]map[string]bool),
		shutdown:       make(chan struct{}),
		daemon:         d,
	}
}

// Creates logger object, and log file reference
func NewLogger() (*log.Logger, *os.File, error) {
	logPath, err := getLogPath()
	if err != nil {
		return nil, nil, err
	}

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

type DaemonManager interface {
	Install() (string, error)
	Remove() (string, error)
	Start() (string, error)
	Stop() (string, error)
	Status() (string, error)
}

// Command details communicated by pipe
type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}
