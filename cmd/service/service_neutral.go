package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
)

// Handles service commands read from pipe/socket connection
func (s *timekeepService) handleConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Println("INFO: Starting to read from connection.")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			s.logger.Printf("ERROR: Failed to unmarshal JSON from line '%s': %s", line, err)
			continue
		}

		cmd.ProcessName = strings.ToLower(cmd.ProcessName)

		switch cmd.Action {
		case "process_start":
			pid := strconv.Itoa(cmd.ProcessID)
			s.createSession(cmd.ProcessName, pid)
			s.logger.Printf("INFO: Called createSession for %s (PID: %s)", cmd.ProcessName, pid)
		case "process_stop":
			pid := strconv.Itoa(cmd.ProcessID)
			s.endSession(cmd.ProcessName, pid)
			s.logger.Printf("INFO: Called endSession for %s (PID: %s)", cmd.ProcessName, pid)
		case "refresh":
			s.refreshProcessMonitor()
			s.logger.Println("INFO: Called refreshProcessMonitor")
		default:
			s.logger.Printf("WARN: Received unknown command action: %s", cmd.Action)
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Printf("ERROR: Error reading from pipe: %s", err)
	}
}

// Stops the currently running process monitoring script, and starts a new one with updated program list
func (s *timekeepService) refreshProcessMonitor() {
	programs, err := s.prRepo.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Printf("ERROR: Failed to get programs: %s", err)
		return
	}

	s.stopProcessMonitor()

	if len(programs) > 0 {
		s.startProcessMonitor(programs)
	}

	s.logger.Printf("INFO: Process monitor refresh with %d programs", len(programs))
}

// Stops the WMI powershell script
func (s *timekeepService) stopProcessMonitor() {
	if s.psProcess != nil {
		_ = s.psProcess.Process.Kill()
		s.psProcess = nil
	}
}

// If no process is running with given name, will create a new active session in memory and database.
// If there is already a process running with given name, new PID will be added to active session's in-memory map
func (s *timekeepService) createSession(processName string, processID string) {
	pidMap, exists := s.activeSessions[processName]

	if exists {
		pidMap[processID] = true
		s.logger.Printf("INFO: Added PID %s to existing session for %s", processID, processName)
	} else {
		s.activeSessions[processName] = make(map[string]bool)
		s.activeSessions[processName][processID] = true

		sessionParams := database.CreateActiveSessionParams{
			ProgramName: processName,
			StartTime:   time.Now(),
		}

		err := s.asRepo.CreateActiveSession(context.Background(), sessionParams)
		if err != nil {
			s.logger.Printf("ERROR: Error creating active session for process: %s", processName)
			return
		}

		s.logger.Printf("INFO: Created new session for %s with PID %s", processName, processID)
	}
}

// Removes PID from memory map of active sessions, if there are still processes running with given name, session will not end.
// If last process for given name ends, the active session is terminated, and session is moved into session history.
func (s *timekeepService) endSession(processName, processID string) {
	session, ok := s.activeSessions[processName] // Make sure there is a valid active session
	if !ok {
		s.logger.Printf("ERROR: Error ending session: No active session for %s", processName)
		return
	}

	_, ok = session[processID] // Make sure the processID is correctly present in session
	if !ok {
		s.logger.Printf("ERROR: Error ending session: PID %s is not present in session map", processID)
		return
	}

	delete(session, processID)

	if len(session) == 0 {
		s.moveSessionToHistory(processName)
		delete(s.activeSessions, processName)
	}
}

// Takes an active session and moves it into session history, ending active status
func (s *timekeepService) moveSessionToHistory(processName string) {
	startTime, err := s.asRepo.GetActiveSession(context.Background(), processName)
	if err != nil {
		s.logger.Printf("ERROR: Error getting active session from database: %s", err)
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
	err = s.hsRepo.AddToSessionHistory(context.Background(), archivedSession)
	if err != nil {
		s.logger.Printf("ERROR: Error creating session history for %s: %s", processName, err)
		return
	}

	err = s.prRepo.UpdateLifetime(context.Background(), database.UpdateLifetimeParams{
		Name:            processName,
		LifetimeSeconds: duration,
	})
	if err != nil {
		s.logger.Printf("ERROR: Error updating lifetime for %s: %s", processName, err)
	}

	err = s.asRepo.RemoveActiveSession(context.Background(), processName)
	if err != nil {
		s.logger.Printf("ERROR: Error removing active session for %s: %s", processName, err)
	}

	s.logger.Printf("INFO: Moved session for %s to history (duration: %d seconds)", processName, duration)
}

// Closes any open log files
func (s *timekeepService) fileCleanup() {
	if s.logFile != nil {
		s.logger.Println("INFO: Closing log file connection")
		s.logFile.Close()
	}
}
