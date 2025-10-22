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
	newCtx, newCancel := context.WithCancel(parent)

	e.mu.Lock()
	oldCancel := e.WakaCancel
	e.WakaCancel = newCancel
	e.mu.Unlock()

	if oldCancel != nil {
		oldCancel()
	}

	logger.Println("INFO: Starting WakaTime heartbeats")

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

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
	}(newCtx)
}

// Send specified heartbeats to WakaTime
func (e *EventController) sendHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) error {
	type item struct{ program, category, project string }
	items := []item{}

	sm.Mu.Lock()
	for p, t := range sm.Programs {
		if len(t.PIDs) > 0 && t.Category != "" {
			items = append(items, item{p, t.Category, t.Project})
		}
	}
	sm.Mu.Unlock()

	for _, it := range items {
		if err := e.sendWakaHeartbeat(ctx, logger, it.program, it.category, it.project); err != nil {
			return err
		}
		logger.Printf("INFO: WakaTime heartbeat sent for %s, category %s", it.program, it.category)
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
		"--category", category,
		"--time", fmt.Sprintf("%d", time.Now().Unix()),
		"--verbose",
		"--write",
	}

	logger.Printf("DEBUG: cli=%s args=%v", cliPath, args)

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, cliPath, args...)
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
	cancel := e.WakaCancel
	e.WakaCancel = nil
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}
