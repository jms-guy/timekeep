//go:build windows

package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/jms-guy/timekeep/internal/database"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const serviceName = "Timekeep"

//go:embed monitor.ps1
var monitorScript string

// Command details communicated by pipe
type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

func RunService(name string, isDebug *bool) error {
	service, err := ServiceSetup()
	if err != nil {
		return err
	}
	if *isDebug {
		return debug.Run(name, service)
	} else {
		return svc.Run(name, service)
	}
}

// Service execute method for Windows Handler interface
func (s *timekeepService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	s.logger.Println("INFO: Service Execute function entered.")

	err := s.logFile.Sync()
	if err != nil {
		s.logger.Printf("ERROR: Failed to sync log file: %v", err)
	}

	// Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	programs, err := s.prRepo.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Printf("ERROR: Failed to get programs: %s", err)
		return false, 1
	}
	if len(programs) > 0 {
		s.startProcessMonitor(programs)
	}

	go s.listenPipe()

	// Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: // Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: // Service needs to be stopped or shutdown
				close(s.shutdown)
				s.fileCleanup()
				break loop
			case svc.Pause: // Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: // Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				s.logger.Printf("ERROR: Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 0
}

// Opens a Windows named pipe connection, to listen for commands
func (s *timekeepService) listenPipe() {
	pipeName := "\\\\.\\pipe\\Timekeep"

	pipe, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		s.logger.Printf("ERROR: Failed to create pipe: %s", err)
		return
	}
	defer pipe.Close()

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := pipe.Accept()
			if err != nil {
				s.logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go s.handlePipeConnection(conn)
		}
	}
}

// Handles service commands read from pipe connection
func (s *timekeepService) handlePipeConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Println("INFO: Starting to read from pipe.")

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

// Runs the powershell WMI script, to monitor process events
func (s *timekeepService) startProcessMonitor(programs []string) {
	programList := strings.Join(programs, ",")

	scriptTempDir := filepath.Join("C:\\", "ProgramData", "TimeKeep", "scripts_temp")

	if err := os.MkdirAll(scriptTempDir, 0o755); err != nil {
		s.logger.Printf("ERROR: Failed to create PowerShell script temp directory '%s': %s", scriptTempDir, err)
		return
	}

	tempFile, err := os.CreateTemp(scriptTempDir, "monitor*.ps1")
	if err != nil {
		s.logger.Printf("ERROR: Failed to create temp script file in '%s': %s", scriptTempDir, err)
		return
	}

	defer tempFile.Close()

	if _, err := tempFile.WriteString(monitorScript); err != nil {
		s.logger.Printf("ERROR: Failed to write script: %s", err)
		return
	}

	if err := tempFile.Sync(); err != nil {
		s.logger.Printf("ERROR: Failed to sync temp script file to disk: %s", err)
		return
	}

	tempFile.Close()

	time.Sleep(100 * time.Millisecond) // Pause to allow tempfile to finish writing before it attempts to execute

	args := []string{"-ExecutionPolicy", "Bypass", "-File", tempFile.Name(), "-Programs", programList}
	cmd := exec.Command("powershell", args...)
	s.psProcess = cmd

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		s.logger.Printf("ERROR: Failed to start PowerShell monitor: %s", err)
		s.psProcess = nil
		if stderr.Len() > 0 {
			s.logger.Printf("INFO: PowerShell stderr (on Start() failure): %s", stderr.String())
		}
	}

	// Goroutine to wait for the PowerShell process to exit and log its stderr/stdout
	go func() {
		defer os.Remove(tempFile.Name())

		err := cmd.Wait()
		if err != nil {
			s.logger.Printf("ERROR: PowerShell monitor process exited with error: %s", err)
		} else {
			s.logger.Println("INFO: PowerShell monitor process exited successfully.")
		}

		if stderr.Len() > 0 {
			s.logger.Printf("PowerShell stderr output: %s", stderr.String())
		} else {
			s.logger.Println("INFO: No PowerShell stderr output.")
		}
	}()
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
			StartTime:   time.Now().UTC(),
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
	endTime := time.Now().UTC()
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
