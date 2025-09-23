//go:build linux

package events

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (e *EventController) MonitorProcesses(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	for {
		livePIDS := e.checkForProcessStartEvents(logger, s, pr, a, programs)
		e.checkForProcessStopEvents(logger, s, pr, a, h, livePIDS)
		time.Sleep(1 * time.Second)
	}
}

func (e *EventController) checkForProcessStartEvents(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, programs []string) map[int]struct{} {
	entries, err := os.ReadDir("/proc") // Read /proc
	if err != nil {
		logger.Printf("ERROR: Couldn't read /proc: %s", err)
		return nil
	}

	live := make(map[int]struct{}, len(entries))
	for _, e := range entries { // Loop over PID entries
		if !e.IsDir() {
			continue
		}
		pid, ok := parsePID(e.Name())
		if !ok {
			continue
		}
		live[pid] = struct{}{}

		identity, err := getProgramIdentity(pid)
		if err != nil {
			continue
		}
		actual := filepath.Base(identity)

		s.Mu.Lock()
		_, match := s.Programs[actual] // Is program being tracked?
		if !match {
			s.Mu.Unlock()
			continue
		}

		tracked := false
		if t := s.Programs[actual]; t != nil {
			if _, exists := t.PIDs[pid]; exists {
				tracked = true
			}
		}
		s.Mu.Unlock()
		if tracked {
			continue
		}

		s.CreateSession(logger, a, actual, pid)
	}

	return live
}

func (e *EventController) checkForProcessStopEvents(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, livePIDs map[int]struct{}) {
	if livePIDs == nil {
		livePIDs = map[int]struct{}{}
	}

	s.Mu.Lock()
	type toEnd struct {
		program string
		pid     int
	}
	var ends []toEnd

	now := time.Now()
	// Loop tracked programs. For each PID currently being tracked, check if it exists in the live map. If it does, update last seen value,
	// else schedule the PID to be removed from tracking
	for program, t := range s.Programs {
		if t == nil {
			continue
		}

		for pid := range t.PIDs {
			if _, ok := livePIDs[pid]; ok {
				t.LastSeen = now
				continue
			}

			ends = append(ends, toEnd{program, pid})
		}
	}
	s.Mu.Unlock()

	for _, eend := range ends {
		s.EndSession(logger, pr, a, h, eend.program, eend.pid)
	}
}

func (e *EventController) StopProcessMonitor() {
}

// Read process /proc/{pid}/exe path to get program name
func readExePath(pid int) (string, error) {
	p := fmt.Sprintf("/proc/%d/exe", pid)
	target, err := os.Readlink(p)
	if err != nil {
		return "", err
	}

	real, err := filepath.EvalSymlinks(target)
	if err != nil {
		return target, nil
	}

	return real, nil
}

// Read process /proc/{pid}/cmdline path to get program name
func readCmdline(pid int) (string, error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}

	parts := strings.Split(string(b), "\x00")
	if len(parts) == 0 || parts[0] == "" {
		return "", fmt.Errorf("empty cmdline")
	}

	return parts[0], nil
}

// Get identity of process by reading exe and cmdline paths
func getProgramIdentity(pid int) (string, error) {
	if exe, err := readExePath(pid); err == nil && exe != "" {
		return exe, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if argv0, err := readCmdline(pid); err == nil && argv0 != "" {
		return argv0, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	// Fallback to /comm
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func parsePID(name string) (int, bool) {
	pid, err := strconv.Atoi(name)
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}
