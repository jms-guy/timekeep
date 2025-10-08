package events

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
)

// Start WakaTime heartbeat ticker
func (e *EventController) StartHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) {
	e.heartbeatMu.Lock()
	e.wakaHeartbeatTicker = time.NewTicker(1 * time.Minute)
	ticker := e.wakaHeartbeatTicker
	e.heartbeatMu.Unlock()

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

				if err := e.sendHeartbeats(ctx, sm); err != nil {
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
func (e *EventController) sendHeartbeats(ctx context.Context, sm *sessions.SessionManager) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	for program, tracked := range sm.Programs {
		if len(tracked.PIDs) > 0 {
			if tracked.Category != "" {
				if err := e.sendWakaHeartbeat(ctx, program, tracked.Category); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Call the wakatime-cli heartbeat command
func (e *EventController) sendWakaHeartbeat(ctx context.Context, program, category string) error {
	cmd := exec.CommandContext(ctx,
		"wakatime-cli",
		"--entity", program,
		"--entity-type", "app",
		"--plugin", "timekeep/"+e.version,
		"--category", category,
		"--time", fmt.Sprintf("%d", time.Now().Unix()))

	return cmd.Run()
}

// Stops WakaTime heartbeat ticker after disabling integration
func (e *EventController) StopHeartbeats() {
	e.heartbeatMu.Lock()
	defer e.heartbeatMu.Unlock()

	if e.wakaHeartbeatTicker != nil {
		e.wakaHeartbeatTicker.Stop()
		e.wakaHeartbeatTicker = nil
	}
}
