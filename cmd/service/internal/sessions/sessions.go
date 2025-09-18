package sessions

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
	"github.com/jms-guy/timekeep/internal/repository"
)

type Tracked struct {
	PIDs     map[int]struct{}
	StartAt  time.Time
	LastSeen time.Time
}

type SessionManager struct {
	Programs map[string]*Tracked
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{Programs: make(map[string]*Tracked)}
}

// Make sure map is initialized, add program to map if not already present
func (sm *SessionManager) EnsureProgram(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.Programs == nil {
		sm.Programs = make(map[string]*Tracked)
	}
	if _, ok := sm.Programs[name]; !ok {
		sm.Programs[name] = &Tracked{PIDs: make(map[int]struct{})}
	}
}

// If no process is running with given name, will create a new active session in database.
// If there is already a process running with given name, new PID will be added to active session
func (sm *SessionManager) CreateSession(logger *log.Logger, a repository.ActiveRepository, processName string, pid int) {
	sm.mu.Lock()

	t := sm.Programs[processName]
	if t == nil {
		t = &Tracked{PIDs: make(map[int]struct{})}
		sm.Programs[processName] = t
	}

	if _, ok := t.PIDs[pid]; ok {
		sm.mu.Unlock()
		logger.Printf("INFO: PID %d already tracked for %s", pid, processName)
		return
	}
	t.PIDs[pid] = struct{}{}

	now := time.Now()
	if len(t.PIDs) == 1 {
		t.StartAt = now
	}

	t.LastSeen = now
	sm.mu.Unlock()

	if len(t.PIDs) == 1 {
		params := database.CreateActiveSessionParams{ProgramName: processName, StartTime: now}
		if err := a.CreateActiveSession(context.Background(), params); err != nil {
			logger.Printf("ERROR: creating active session for %s: %v", processName, err)
			return
		}
		logger.Printf("INFO: Created new session for %s at %s", processName, now)
	} else {
		logger.Printf("INFO: Added PID %d to existing session for %s", pid, processName)
	}
}

// Removes PID from sessions map, if there are still processes running with given name, session will not end.
// If last process for given name ends, the active session is terminated, and session is moved into session history.
func (sm *SessionManager) EndSession(logger *log.Logger, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, processName string, pid int) {
	sm.mu.Lock()

	t, ok := sm.Programs[processName]
	if !ok {
		sm.mu.Unlock()
		logger.Printf("INFO: No active session for %s (pid %d)", processName, pid)
		return
	}

	if _, ok := t.PIDs[pid]; !ok {
		sm.mu.Unlock()
		logger.Printf("INFO: PID %d not tracked for %s", pid, processName)
		return
	}

	delete(t.PIDs, pid)

	now := time.Now()
	t.LastSeen = now
	sm.mu.Unlock()

	if len(t.PIDs) == 0 {
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
