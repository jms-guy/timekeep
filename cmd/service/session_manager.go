package main

import (
	"context"
	"log"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
	"github.com/jms-guy/timekeep/internal/repository"
)

type SessionManager struct {
	Programs map[string]map[int]struct{} // Map of programs being tracked
}

func NewSessionManager() *SessionManager {
	return &SessionManager{Programs: make(map[string]map[int]struct{})}
}

// If no process is running with given name, will create a new active session in memory and database.
// If there is already a process running with given name, new PID will be added to active session's in-memory map
func (sm *SessionManager) CreateSession(logger *log.Logger, a repository.ActiveRepository, processName string, processID string) {
	pidMap, exists := sm.Programs[processName]

	if exists {
		pidMap[processID] = true
		logger.Printf("INFO: Added PID %s to existing session for %s", processID, processName)
	} else {
		sm.Programs[processName] = make(map[string]bool)
		sm.Programs[processName][processID] = true

		sessionParams := database.CreateActiveSessionParams{
			ProgramName: processName,
			StartTime:   time.Now(),
		}

		err := a.CreateActiveSession(context.Background(), sessionParams)
		if err != nil {
			logger.Printf("ERROR: Error creating active session for process: %s", processName)
			return
		}

		logger.Printf("INFO: Created new session for %s with PID %s", processName, processID)
	}
}

// Removes PID from memory map of active sessions, if there are still processes running with given name, session will not end.
// If last process for given name ends, the active session is terminated, and session is moved into session history.
func (sm *SessionManager) EndSession(logger *log.Logger, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, processName, processID string) {
	session, ok := s.programs[processName] // Make sure there is a valid active session
	if !ok {
		logger.Printf("ERROR: Error ending session: No active session for %s", processName)
		return
	}

	_, ok = session[processID] // Make sure the processID is correctly present in session
	if !ok {
		logger.Printf("ERROR: Error ending session: PID %s is not present in session map", processID)
		return
	}

	delete(session, processID)

	if len(session) == 0 {
		sm.MoveSessionToHistory(logger, pr, a, h, processName)
		delete(sm.Programs, processName)
	}
}

// Takes an active session and moves it into session history, ending active status
func (sm *SessionManager) MoveSessionToHistory(logger *log.Logger, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, processName string) {
	startTime, err := a.GetActiveSession(context.Background(), processName)
	if err != nil {
		logger.Printf("ERROR: Error getting active session from database: %s", err)
		return
	}
	endTime := time.Now()
	duration := int64(endTime.Sub(startTime).Seconds())

	archivedSession := database.AddToSessionHistoryParams{
		ProgramName:     processName,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: duration,
	}
	err = h.AddToSessionHistory(context.Background(), archivedSession)
	if err != nil {
		logger.Printf("ERROR: Error creating session history for %s: %s", processName, err)
		return
	}

	err = pr.UpdateLifetime(context.Background(), database.UpdateLifetimeParams{
		Name:            processName,
		LifetimeSeconds: duration,
	})
	if err != nil {
		logger.Printf("ERROR: Error updating lifetime for %s: %s", processName, err)
	}

	err = a.RemoveActiveSession(context.Background(), processName)
	if err != nil {
		logger.Printf("ERROR: Error removing active session for %s: %s", processName, err)
	}

	logger.Printf("INFO: Moved session for %s to history (duration: %d seconds)", processName, duration)
}
