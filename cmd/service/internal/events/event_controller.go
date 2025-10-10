package events

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/config"
	"github.com/jms-guy/timekeep/internal/repository"
)

var Version = "dev"

// Command details communicated by pipe
type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

type EventController struct {
	PsProcess           *exec.Cmd          // Powershell process for Windows event monitoring
	cancel              context.CancelFunc // Event monitoring cancel context
	Config              *config.Config     // Struct built from config file
	wakaHeartbeatTicker *time.Ticker       // Ticker for WakaTime enabled heartbeats
	heartbeatMu         sync.Mutex         // Mutex for WakaTime heartbeat ticker
	version             string             // Timekeep version
}

func NewEventController() *EventController {
	return &EventController{version: Version}
}

// Handles service commands read from pipe/socket connection
func (e *EventController) HandleConnection(serviceCtx context.Context, logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, conn net.Conn) {
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

		cmdCtx, cancel := context.WithTimeout(serviceCtx, 5*time.Second)

		switch cmd.Action {
		case "process_start":
			s.CreateSession(cmdCtx, logger, a, cmd.ProcessName, cmd.ProcessID)
			logger.Printf("INFO: Called createSession for %s (PID: %d)", cmd.ProcessName, cmd.ProcessID)
		case "process_stop":
			s.EndSession(cmdCtx, logger, pr, a, h, cmd.ProcessName, cmd.ProcessID)
			logger.Printf("INFO: Called endSession for %s (PID: %d)", cmd.ProcessName, cmd.ProcessID)
		case "refresh":
			e.RefreshProcessMonitor(cmdCtx, logger, s, pr, a, h)
			logger.Println("INFO: Called refreshProcessMonitor")
		default:
			logger.Printf("WARN: Received unknown command action: %s", cmd.Action)
		}

		cancel()
	}

	if err := scanner.Err(); err != nil {
		logger.Printf("ERROR: Error reading from pipe: %s", err)
	}
}

// Stops the currently running process monitoring script, and starts a new one with updated program list
func (e *EventController) RefreshProcessMonitor(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	e.StopProcessMonitor()

	programs, err := pr.GetAllPrograms(ctx)
	if err != nil {
		logger.Printf("ERROR: Failed to get programs: %s", err)
		return
	}

	if len(programs) > 0 {
		toTrack := []string{}
		for _, program := range programs {
			category := ""
			if program.Category.Valid {
				category = program.Category.String
			}
			sm.EnsureProgram(program.Name, category)

			toTrack = append(toTrack, program.Name)
		}

		go e.MonitorProcesses(ctx, logger, sm, pr, a, h, toTrack)
	}

	newConfig, err := config.Load()
	if err != nil {
		logger.Printf("ERROR: Failed to load config: %s", err)
		return
	}

	if newConfig.WakaTime.Enabled && !e.Config.WakaTime.Enabled {
		e.StartHeartbeats(ctx, logger, sm)
	} else if !newConfig.WakaTime.Enabled && e.Config.WakaTime.Enabled {
		e.StopHeartbeats()
	}

	e.Config = newConfig

	logger.Printf("INFO: Process monitor refresh with %d programs", len(programs))
}
