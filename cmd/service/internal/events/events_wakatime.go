package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
)

// Start WakaTime heartbeat ticker
func (e *EventController) StartHeartbeats(parent context.Context, logger *log.Logger, sm *sessions.SessionManager) {
	e.mu.Lock()
	if e.WakaCancel != nil {
		e.WakaCancel()
		e.WakaCancel = nil
	}
	ctx, cancel := context.WithCancel(parent)
	e.WakaCancel = cancel
	e.mu.Unlock()

	e.heartbeatMu.Lock()
	if e.wakaHeartbeatTicker != nil {
		e.wakaHeartbeatTicker.Stop()
	}
	e.wakaHeartbeatTicker = time.NewTicker(time.Minute)
	ticker := e.wakaHeartbeatTicker
	e.heartbeatMu.Unlock()

	logger.Println("INFO: Starting WakaTime heartbeats")
	go func() {
		defer func() {
			e.heartbeatMu.Lock()
			if e.wakaHeartbeatTicker != nil {
				e.wakaHeartbeatTicker.Stop()
				e.wakaHeartbeatTicker = nil
			}
			e.heartbeatMu.Unlock()
		}()
		errorCount := 0
		for {
			select {
			case <-ctx.Done():
				logger.Println("INFO: Stopping WakaTime heartbeats")
				return
			case <-ticker.C:
				if errorCount >= 5 {
					logger.Println("ERROR: WakaTime heartbeats failed 5 times consecutively, stopping")
					return
				}
				if err := e.sendHeartbeats(ctx, logger, sm); err != nil {
					logger.Printf("ERROR: Failed to send WakaTime heartbeat: %s", err)
					errorCount++
					continue
				}
				errorCount = 0
			}
		}
	}()
}

// Send specified heartbeats to WakaTime
func (e *EventController) sendHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	for program, tracked := range sm.Programs {
		if len(tracked.PIDs) > 0 {
			if tracked.Category != "" {
				if err := e.sendWakaHeartbeat(ctx, logger, program, tracked.Category, tracked.Project); err != nil {
					return err
				}
				logger.Printf("INFO: WakaTime heartbeat sent for %s, category %s", program, tracked.Category)
				continue
			}
			logger.Printf("INFO: WakaTime heartbeat skipped for %s, no category set", program)
		}
	}

	return nil
}

// Call the wakatime-cli heartbeat command
func (e *EventController) sendWakaHeartbeat(ctx context.Context, logger *log.Logger, program, category, project string) error {
	cliPath := e.Config.WakaTime.CLIPath

	if cliPath == "" {
		logger.Println("ERROR: wakatime-cli path not set")
	}

	projectToUse := e.Config.WakaTime.GlobalProject
	if project != "" {
		projectToUse = project
	}
	logger.Printf("INFO: Sending WakaTime heartbeat for %s, category %s, project %s", program, category, projectToUse)

	args := []string{
		"--key", e.Config.WakaTime.APIKey,
		"--entity", program,
		"--entity-type", "app",
		"--plugin", "timekeep/" + e.version,
		"--alternate-project", projectToUse,
		"--category", category,
		"--time", fmt.Sprintf("%d", time.Now().Unix()),
		"--verbose",
	}

	cmd := exec.CommandContext(ctx, cliPath, args...)
	cmd.Env = append(os.Environ(),
		"HOME=/home/jamieguy",
		"PATH=/usr/local/bin:/usr/bin",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("ERROR: wakatime-cli failed: %v, output: %s", err, out)
		return err
	}
	return nil
}

// Stops WakaTime heartbeat ticker after disabling integration
func (e *EventController) StopHeartbeats() {
	e.mu.Lock()
	if e.WakaCancel != nil {
		e.WakaCancel()
		e.WakaCancel = nil
	}
	e.mu.Unlock()

	e.heartbeatMu.Lock()
	if e.wakaHeartbeatTicker != nil {
		e.wakaHeartbeatTicker.Stop()
		e.wakaHeartbeatTicker = nil
	}
	e.heartbeatMu.Unlock()
}
