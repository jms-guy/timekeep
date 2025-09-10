//go:build windows

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/jms-guy/timekeep/internal/database"
	mysql "github.com/jms-guy/timekeep/sql"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const serviceName = "Timekeep"

//go:embed monitor.ps1
var monitorScript string

// Service context
type timekeepService struct {
	db             *database.Queries          // SQLC database queries
	logger         *log.Logger                // Logging object
	logFile        *os.File                   // Reference to the output log file
	psProcess      *exec.Cmd                  // The running WMI powershell script
	activeSessions map[string]map[string]bool // Map of active sessions & their PIDs
	shutdown       chan struct{}              // Shutdown channel
}

// Command details communicated by pipe
type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

// Creates new service instance
func NewTimekeepService() (*timekeepService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	logPath, err := getLogPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := log.New(f, "Timekeep: ", log.LstdFlags)

	return &timekeepService{
		db:             db,
		logger:         logger,
		logFile:        f,
		activeSessions: make(map[string]map[string]bool),
		shutdown:       make(chan struct{}),
	}, nil
}

func getLogPath() (string, error) {
	logDir := `C:\ProgramData\TimeKeep\logs`
	return filepath.Join(logDir, "timekeep.log"), nil
}

func RunService(name string, isDebug bool) error {
	service, err := NewTimekeepService()
	if err != nil {
		return err
	}
	if isDebug {
		return debug.Run(name, service)
	} else {
		return svc.Run(name, service)
	}
}

// Service execute method for Windows Handler interface
func (s *timekeepService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {

	//Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	programs, err := s.db.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Printf("Failed to get programs: %s", err)
		return false, 1
	}
	if len(programs) > 0 {
		s.startProcessMonitor(programs)
	}

	go s.listenPipe()

	//Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: //Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: //Service needs to be stopped or shutdown
				close(s.shutdown)
				s.fileCleanup()
				break loop
			case svc.Pause: //Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: //Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				s.logger.Printf("Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

// Opens a Windows named pipe connection, to listen for commands
func (s *timekeepService) listenPipe() {
	pipeName := `\\.\pipe\Timekeep`

	pipe, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		s.logger.Printf("Failed to create pipe: %s", err)
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
				s.logger.Printf("Failed to accept connection: %s", err)
				continue
			}
			go s.handlePipeConnection(conn)
		}
	}
}

// Handles service commands read from pipe connection
func (s *timekeepService) handlePipeConnection(conn net.Conn) {
	defer conn.Close()

	var cmd Command
	decoder := json.NewDecoder(conn)

	if err := decoder.Decode(&cmd); err != nil {
		s.logger.Printf("Failed to decode command: %s", err)
		return
	}

	switch cmd.Action {
	case "process_start":
		pid := strconv.Itoa(cmd.ProcessID)
		s.createSession(cmd.ProcessName, pid)
	case "process_stop":
		pid := strconv.Itoa(cmd.ProcessID)
		s.endSession(cmd.ProcessName, pid)
	case "refresh":
		s.refreshProcessMonitor()
	}
}

// Runs the powershell WMI script, to monitor process events
func (s *timekeepService) startProcessMonitor(programs []string) {
	programList := strings.Join(programs, ",")

	tempFile, err := os.CreateTemp("", "monitor*.ps1")
	if err != nil {
		s.logger.Printf("Failed to create temp script file: %s", err)
		return
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(monitorScript); err != nil {
		s.logger.Printf("Failed to write script: %s", err)
		return
	}

	cmd := exec.Command("powershell", "-File", tempFile.Name(), "-Programs", programList)
	s.psProcess = cmd

	if err := cmd.Start(); err != nil {
		s.logger.Printf("Failed to start PowerShell monitor: %s", err)
		s.psProcess = nil
	}
}

// Stops the currently running process monitoring script, and starts a new one with updated program list
func (s *timekeepService) refreshProcessMonitor() {
	programs, err := s.db.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Printf("Failed to get programs: %s", err)
		return
	}

	s.stopProcessMonitor()

	if len(programs) > 0 {
		s.startProcessMonitor(programs)
	}

	s.logger.Printf("Process monitor refresh with %d programs", len(programs))
}

// Stops the WMI powershell script
func (s *timekeepService) stopProcessMonitor() {
	if s.psProcess != nil {
		s.psProcess.Process.Signal(os.Interrupt)
		s.psProcess = nil
	}
}

// If no process is running with given name, will create a new active session in memory and database.
// If there is already a process running with given name, new PID will be added to active session's in-memory map
func (s *timekeepService) createSession(processName string, processID string) {
	pidMap, exists := s.activeSessions[processName]

	if exists {
		pidMap[processID] = true
		s.logger.Printf("Added PID %s to existing session for %s", processID, processName)
	} else {
		s.activeSessions[processName] = make(map[string]bool)
		s.activeSessions[processName][processID] = true

		sessionParams := database.CreateActiveSessionParams{
			ProgramName: processName,
			StartTime:   time.Now().UTC(),
		}

		err := s.db.CreateActiveSession(context.Background(), sessionParams)
		if err != nil {
			s.logger.Printf("Error creating active session for process: %s", processName)
			return
		}

		s.logger.Printf("Created new session for %s with PID %s", processName, processID)
	}
}

// Removes PID from memory map of active sessions, if there are still processes running with given name, session will not end.
// If last process for given name ends, the active session is terminated, and session is moved into session history.
func (s *timekeepService) endSession(processName, processID string) {
	session, ok := s.activeSessions[processName] // Make sure there is a valid active session
	if !ok {
		s.logger.Printf("Error ending session: No active session for %s", processName)
		return
	}

	_, ok = session[processID] // Make sure the processID is correctly present in session
	if !ok {
		s.logger.Printf("Error ending session: PID %s is not present in session map", processID)
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
	startTime, err := s.db.GetActiveSession(context.Background(), processName)
	if err != nil {
		s.logger.Printf("Error getting active session from database: %s", err)
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
	err = s.db.AddToSessionHistory(context.Background(), archivedSession)
	if err != nil {
		s.logger.Printf("Error creating session history for %s: %s", processName, err)
		return
	}

	err = s.db.UpdateLifetime(context.Background(), database.UpdateLifetimeParams{
		Name:            processName,
		LifetimeSeconds: duration,
	})
	if err != nil {
		s.logger.Printf("Error updating lifetime for %s: %s", processName, err)
	}

	err = s.db.RemoveActiveSession(context.Background(), processName)
	if err != nil {
		s.logger.Printf("Error removing active session for %s: %s", processName, err)
	}

	s.logger.Printf("Moved session for %s to history (duration: %d seconds)", processName, duration)
}

// Closes any open log files
func (s *timekeepService) fileCleanup() {
	if s.logFile != nil {
		s.logFile.Close()
	}
}
