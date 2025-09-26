package events

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

// Command details communicated by pipe
type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

type EventController struct {
	PsProcess *exec.Cmd // Powershell process for Windows event monitoring
	cancel    context.CancelFunc
}

func NewEventController() *EventController {
	return &EventController{}
}

// Handles service commands read from pipe/socket connection
func (e *EventController) HandleConnection(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, conn net.Conn) {
	defer conn.Close()

	logger.Println("INFO: Starting to read from connection.")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			logger.Printf("ERROR: Failed to unmarshal JSON from line '%s': %s", line, err)
			continue
		}

		cmd.ProcessName = strings.ToLower(cmd.ProcessName)

		switch cmd.Action {
		case "process_start":
			s.CreateSession(logger, a, cmd.ProcessName, cmd.ProcessID)
			logger.Printf("INFO: Called createSession for %s (PID: %d)", cmd.ProcessName, cmd.ProcessID)
		case "process_stop":
			s.EndSession(logger, pr, a, h, cmd.ProcessName, cmd.ProcessID)
			logger.Printf("INFO: Called endSession for %s (PID: %d)", cmd.ProcessName, cmd.ProcessID)
		case "refresh":
			e.RefreshProcessMonitor(logger, s, pr, a, h)
			logger.Println("INFO: Called refreshProcessMonitor")
		default:
			logger.Printf("WARN: Received unknown command action: %s", cmd.Action)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Printf("ERROR: Error reading from pipe: %s", err)
	}
}

// Stops the currently running process monitoring script, and starts a new one with updated program list
func (e *EventController) RefreshProcessMonitor(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	programs, err := pr.GetAllProgramNames(context.Background())
	if err != nil {
		logger.Printf("ERROR: Failed to get programs: %s", err)
		return
	}

	e.StopProcessMonitor()

	if len(programs) > 0 {
		for _, program := range programs {
			s.EnsureProgram(program)
		}
		go e.MonitorProcesses(logger, s, pr, a, h, programs)
	}

	logger.Printf("INFO: Process monitor refresh with %d programs", len(programs))
}
