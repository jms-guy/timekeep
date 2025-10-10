package main

import (
	"context"

	"github.com/jms-guy/timekeep/cmd/service/internal/daemons"
	"github.com/jms-guy/timekeep/cmd/service/internal/events"
	"github.com/jms-guy/timekeep/cmd/service/internal/logs"
	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/cmd/service/internal/transport"
	"github.com/jms-guy/timekeep/internal/config"
	"github.com/jms-guy/timekeep/internal/repository"
	mysql "github.com/jms-guy/timekeep/sql"
)

// Service context
type timekeepService struct {
	prRepo    repository.ProgramRepository // Repository for tracked_programs database queries
	asRepo    repository.ActiveRepository  // Repository for active_sessions database queries
	hsRepo    repository.HistoryRepository // Repository for session_history database queries
	logger    *logs.Logs                   // Handles logging operations
	eventCtrl *events.EventController      // Managing struct of OS-specific process monitoring functions & handling transport connection events
	sessions  *sessions.SessionManager     // Managing struct for program sessions
	transport *transport.Transporter       // Handles receiving pipe/socket commands & events
	daemon    daemons.DaemonManager        // Embedded daemon.Daemon struct wrapped by interface
}

func ServiceSetup() (*timekeepService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	logger, err := logs.NewLogs()
	if err != nil {
		return nil, err
	}

	d, err := daemons.NewDaemonManager()
	if err != nil {
		return nil, err
	}

	eventCtrl := events.NewEventController()
	sessions := sessions.NewSessionManager()
	ts := transport.NewTransporter()

	service := NewTimekeepService(store, store, store, logger, eventCtrl, sessions, ts, d)

	config, err := config.Load()
	if err != nil {
		return nil, err
	}

	service.eventCtrl.Config = config

	return service, nil
}

func TestServiceSetup() (*timekeepService, error) {
	db, err := mysql.OpenTestDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	logger := logs.NewTestLogs()

	sessions := sessions.NewSessionManager()

	service := NewTimekeepService(store, store, store, logger, nil, sessions, nil, nil)

	return service, nil
}

// Creates new service instance
func NewTimekeepService(pr repository.ProgramRepository, ar repository.ActiveRepository, hr repository.HistoryRepository, logger *logs.Logs, eventCtrl *events.EventController, sessions *sessions.SessionManager, ts *transport.Transporter, d daemons.DaemonManager) *timekeepService {
	return &timekeepService{
		prRepo:    pr,
		asRepo:    ar,
		hsRepo:    hr,
		logger:    logger,
		eventCtrl: eventCtrl,
		sessions:  sessions,
		transport: ts,
		daemon:    d,
	}
}

// Service shutdown function
func (s *timekeepService) closeService(ctx context.Context) {
	if s.eventCtrl.Config.WakaTime.Enabled { // Stop WakaTime heartbeats
		s.eventCtrl.StopHeartbeats()
	}
	s.eventCtrl.StopProcessMonitor() // Stop any current monitoring function
	s.logger.FileCleanup()           // Close open logging file

	s.sessions.Mu.Lock()
	for program, tracked := range s.sessions.Programs { // End any active sessions
		if len(tracked.PIDs) != 0 {
			s.sessions.MoveSessionToHistory(ctx, s.logger.Logger, s.prRepo, s.asRepo, s.hsRepo, program)
		}
	}
	s.sessions.Mu.Unlock()
}
