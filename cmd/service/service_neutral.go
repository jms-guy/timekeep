package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"
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
			s.Sessions.CreateSession(s.logger, s.asRepo, cmd.ProcessName, pid)
			s.logger.Printf("INFO: Called createSession for %s (PID: %s)", cmd.ProcessName, pid)
		case "process_stop":
			pid := strconv.Itoa(cmd.ProcessID)
			s.Sessions.EndSession(s.logger, s.prRepo, s.asRepo, s.hsRepo, cmd.ProcessName, pid)
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

// Closes any open log files
func (s *timekeepService) fileCleanup() {
	if s.logFile != nil {
		s.logger.Println("INFO: Closing log file connection")
		s.logFile.Close()
	}
}
