//go:build linux

package events

import (
	"context"
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

// Linux specific event functions, handling PID tracking through /proc polling

func (e *EventController) StartMonitor(parent context.Context, logger *log.Logger, sm *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	e.mu.Lock()
	if e.MonCancel != nil {
		e.MonCancel()
		e.MonCancel = nil
	}
	ctx, cancel := context.WithCancel(parent)
	e.MonCancel = cancel
	e.mu.Unlock()

	go e.MonitorProcesses(ctx, logger, sm, pr, a, h, programs)
}

// Main process monitoring function for Linux version
func (e *EventController) MonitorProcesses(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	logger.Println("INFO: Executing main process monitor")

	pollInterval := e.pollTime()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Grace period for PID tracking, to allow for accidently missed PIDs while polling
	if e.Config.PollGrace <= 0 {
		e.Config.PollGrace = 3
	}
	grace := pollInterval * time.Duration(e.Config.PollGrace)

	for {
		select {
		case <-ctx.Done():
			logger.Println("INFO: Monitor context cancelled")
			return
		case <-ticker.C:
			livePIDS := e.checkForProcessStartEvents(logger, sm, a)
			e.checkForProcessStopEvents(logger, sm, pr, a, h, livePIDS, grace)
		}
	}
}

// Polls /proc and loops over PID entries, looking for any new PIDS belonging to tracked programs
func (e *EventController) checkForProcessStartEvents(logger *log.Logger, sm *sessions.SessionManager, a repository.ActiveRepository) map[int]struct{} {
	entries, err := os.ReadDir("/proc") // Read /proc
	if err != nil {
		logger.Printf("ERROR: Couldn't read /proc: %s", err)
		return nil
	}

	live := make(map[int]struct{})
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

		sm.Mu.Lock()
		_, match := sm.Programs[identity] // Is program being tracked?
		if !match {
			sm.Mu.Unlock()
			continue
		}

		if t := sm.Programs[identity]; t != nil {
			if _, exists := t.PIDs[pid]; exists {
				t.LastSeen = time.Now()
				sm.Mu.Unlock()
				continue
			}
		}
		sm.Mu.Unlock()

		sm.CreateSession(context.Background(), logger, a, identity, pid)
	}

	return live
}

// Takes the PID entries found in the previous check function, and compares them against map of active PIDs, to determine if
// any active sessions need ending
func (e *EventController) checkForProcessStopEvents(logger *log.Logger, sm *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, livePIDs map[int]struct{}, grace time.Duration) {
	if livePIDs == nil {
		livePIDs = map[int]struct{}{}
	}

	sm.Mu.Lock()
	type toEnd struct {
		program string
		pid     int
	}
	var ends []toEnd

	now := time.Now()
	// Loop tracked programs. For each PID currently being tracked, check if it exists in the live map. If it does, update last seen value,
	// else schedule the PID to be removed from tracking
	for program, t := range sm.Programs {
		if t == nil {
			continue
		}

		for pid := range t.PIDs {
			if _, ok := livePIDs[pid]; ok {
				t.LastSeen = now
				continue
			}

			if now.Sub(t.LastSeen) >= grace {
				ends = append(ends, toEnd{program, pid})
			}
		}
	}
	sm.Mu.Unlock()

	for _, eend := range ends {
		sm.EndSession(context.Background(), logger, pr, a, h, eend.program, eend.pid)
	}
}

func (e *EventController) StopProcessMonitor() {
	e.mu.Lock()
	if e.MonCancel != nil {
		e.MonCancel()
		e.MonCancel = nil
	}
	e.mu.Unlock()
}

// Determine the polling interval of the monitoring process through config value - defaults to 1s
func (e *EventController) pollTime() time.Duration {
	if e.Config.PollInterval == "" {
		return 1 * time.Second
	}
	d, err := time.ParseDuration(e.Config.PollInterval)
	if err != nil || d <= 0 {
		return 1 * time.Second
	}

	return d
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
		return normalizeBase(exe), nil
	} else if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
		return "", err
	}
	if argv0, err := readCmdline(pid); err == nil && argv0 != "" {
		return normalizeBase(argv0), nil
	} else if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
		return "", err
	}
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}
	return normalizeBase(strings.TrimSpace(string(b))), nil
}

func normalizeBase(s string) string {
	return strings.ToLower(filepath.Base(s))
}

func parsePID(name string) (int, bool) {
	pid, err := strconv.Atoi(name)
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}
