package sessions

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
	"github.com/jms-guy/timekeep/internal/repository"
)

type SessionManager struct {
	Programs map[string]map[string]struct{} // Map of programs being tracked
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{Programs: make(map[string]map[string]struct{})}
}

// Make sure map is initialized, add program to map if not already present
func (sm *SessionManager) EnsureProgram(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.Programs == nil {
		sm.Programs = make(map[string]map[string]struct{})
	}
	if _, ok := sm.Programs[name]; !ok {
		sm.Programs[name] = make(map[string]struct{})
	}
}

// If no process is running with given name, will create a new active session in database.
// If there is already a process running with given name, new PID will be added to active session
func (sm *SessionManager) CreateSession(logger *log.Logger, a repository.ActiveRepository, processName string, processID string) {
	sm.mu.Lock()

	pidMap, exists := sm.Programs[processName]
	if !exists {
		pidMap = make(map[string]struct{})
		sm.Programs[processName] = pidMap
	}
	if _, ok := pidMap[processID]; ok {
		sm.mu.Unlock()
		logger.Printf("INFO: PID %s already tracked for %s", processID, processName)
		return
	}
	pidMap[processID] = struct{}{}
	sm.mu.Unlock()

	if len(pidMap) == 1 {
		params := database.CreateActiveSessionParams{
			ProgramName: processName,
			StartTime:   time.Now(),
		}
		if err := a.CreateActiveSession(context.Background(), params); err != nil {
			logger.Printf("ERROR: creating active session for %s: %v", processName, err)
			return
		}
		logger.Printf("INFO: Created new session for %s at %s", processName, time.Now())
	} else {
		logger.Printf("INFO: Added PID %s to existing session for %s", processID, processName)
	}
}

// Removes PID from sessions map, if there are still processes running with given name, session will not end.
// If last process for given name ends, the active session is terminated, and session is moved into session history.
func (sm *SessionManager) EndSession(logger *log.Logger, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, processName, processID string) {
	sm.mu.Lock()

	pidMap, ok := sm.Programs[processName] // Make sure there is a valid session
	if !ok {
		sm.mu.Unlock()
		logger.Printf("ERROR: Error ending session: No active session for %s", processName)
		return
	}

	_, ok = pidMap[processID] // Make sure the processID is correctly present in session
	if !ok {
		sm.mu.Unlock()
		logger.Printf("ERROR: Error ending session: PID %s is not present in session map", processID)
		return
	}

	delete(pidMap, processID)
	sm.mu.Unlock()

	if len(pidMap) == 0 {
		sm.MoveSessionToHistory(logger, pr, a, h, processName)
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
