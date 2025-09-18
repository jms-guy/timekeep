//go:build linux

package events

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (e *EventController) MonitorProcesses(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	for {
		e.checkForProcessStartEvents(logger, s, pr, a, programs)
		e.checkForProcessStopEvents(logger, s, pr, a, h, programs)
		time.Sleep(1 * time.Second)
	}
}

func (e *EventController) checkForProcessStartEvents(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, programs []string) {
	entries, err := os.ReadDir("/proc") // Read /proc
	if err != nil {
		logger.Printf("ERROR: Couldn't read /proc: %s", err)
		return
	}

	for _, e := range entries { // Loop over PID entries
		if !e.IsDir() {
			continue
		}
		pid := e.Name()

		commPath := filepath.Join("/proc", pid, "comm") // Read /comm file for program name
		b, err := os.ReadFile(commPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				logger.Printf("read %s: %v", commPath, err)
			}
			continue
		}
		comm := strings.TrimSpace(string(b))

		_, ok := s.Programs[comm] // Is program being tracked?
		if !ok {
			continue
		}

		s.CreateSession(logger, a, comm, pid)
	}
}

func (e *EventController) checkForProcessStopEvents(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	currentlyTracked := make(map[string]string) // Remap currently tracked PIDS ({"PID":"Program"})
	for pidKey, _ := range s.Programs {
		key := strings.Split(pidKey, ":")
		currentlyTracked[key[1]] = key[0]
	}

	entries, err := os.ReadDir("/proc") // Read /proc
	if err != nil {
		logger.Printf("ERROR: Couldn't read /proc: %s", err)
		return
	}

	// Loop over /proc entries. If entry is PID directory, check for PID in map. If present in map, check whether name matches the tracked program name.
	// If it matches, remove PID from the map. At end of loop, if any PIDs remain in currently tracked, they will be removed from tracking
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid := e.Name()

		name, ok := currentlyTracked[pid]
		if ok {
			commPath := filepath.Join("/proc", pid, "comm") // Read /comm file for program name
			b, err := os.ReadFile(commPath)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					logger.Printf("read %s: %v", commPath, err)
				}
				continue
			}
			comm := strings.TrimSpace(string(b))

			if comm == name {
				delete(currentlyTracked, pid)
			}
		}
	}

	if len(currentlyTracked) == 0 {
		return
	}

	for processID, processName := range currentlyTracked {
		s.EndSession(logger, pr, a, h, processName, processID)
	}
}

func (e *EventController) StopProcessMonitor() {
	return
}
