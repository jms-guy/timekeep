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
	"github.com/jms-guy/timekeep/internal/database"
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
	PsProcess           *exec.Cmd // Powershell process for Windows event monitoring
	RunCtx              context.Context
	Cancel              context.CancelFunc // Event monitoring cancel context
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
			e.RefreshProcessMonitor(serviceCtx, logger, s, pr, a, h)
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
	logger.Printf("DEBUG: Refresh: Incoming context: %v", ctx)

	logger.Println("DEBUG: Refresh: Stopping heartbeats")
	e.StopHeartbeats()

	logger.Println("DEBUG: Refresh: Stopping process monitor")
	e.StopProcessMonitor()

	if e.Cancel != nil {
		logger.Println("DEBUG: Refresh: Cancelling old context")
		e.Cancel()
	}
	runCtx, runCancel := context.WithCancel(ctx)
	logger.Printf("DEBUG: Refresh: Created new runCtx: %v", runCtx)
	e.RunCtx = runCtx
	e.Cancel = runCancel

	newConfig, err := config.Load()
	if err != nil {
		logger.Printf("ERROR: Failed to load config: %s", err)
		return
	}

	e.Config = newConfig

	logger.Println("DEBUG: Refresh: Getting programs")
	programs, err := pr.GetAllPrograms(context.Background())
	if err != nil {
		logger.Printf("ERROR: Failed to get programs: %s", err)
		return
	}

	logger.Println("DEBUG: Refresh: Updating sessions")
	if len(programs) > 0 {
		toTrack := updateSessionsMapOnRefresh(logger, sm, programs)
		logger.Printf("DEBUG: Refresh: updateSessionsMapOnRefresh returned %d programs to track", len(toTrack))

		logger.Println("DEBUG: Refresh: Starting process monitor")
		go e.MonitorProcesses(e.RunCtx, logger, sm, pr, a, h, toTrack)
	}

	if e.Config.WakaTime.Enabled {
		logger.Println("DEBUG: Refresh: Starting heartbeats")
		e.StartHeartbeats(e.RunCtx, logger, sm)
	}

	logger.Printf("INFO: Process monitor refresh with %d programs", len(programs))
}

// Takes list of programs from database, and updates session map by adding/removing/altering based on any changes from last database grab
func updateSessionsMapOnRefresh(logger *log.Logger, sm *sessions.SessionManager, programs []database.TrackedProgram) []string {
	desired := make(map[string]struct{}, len(programs))
	toTrack := make([]string, 0, len(programs))

	logger.Printf("DEBUG: updateSessionsMapOnRefresh called with %d programs", len(programs))

	sm.Mu.Lock()
	currentKeys := make([]string, 0, len(sm.Programs))
	for k := range sm.Programs {
		currentKeys = append(currentKeys, k)
	}

	logger.Printf("DEBUG: Current session map has %d programs", len(sm.Programs))

	for _, p := range programs {
		name := p.Name
		cat := ""
		if p.Category.Valid {
			cat = p.Category.String
		}
		proj := ""
		if p.Project.Valid {
			proj = p.Project.String
		}

		logger.Printf("DEBUG: Processing program: %s", name)
		sm.EnsureProgram(name, cat, proj)
		desired[name] = struct{}{}
		toTrack = append(toTrack, name)
	}

	if len(desired) == 0 {
		for _, k := range currentKeys {
			delete(sm.Programs, k)
		}
	} else {
		for _, k := range currentKeys {
			if _, keep := desired[k]; !keep {
				delete(sm.Programs, k)
			}
		}
	}
	sm.Mu.Unlock()

	logger.Printf("DEBUG: Returning %d programs to track", len(toTrack))
	return toTrack
}
