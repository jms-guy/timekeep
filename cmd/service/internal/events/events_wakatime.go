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
	e.wakaHeartbeatTicker = time.NewTicker(1 * time.Minute)

	go func() {
		defer e.wakaHeartbeatTicker.Stop()

		errorCount := 0
		for {
			select {
			case <-ctx.Done():
				logger.Println("INFO: Stopping WakaTime heartbeats")
				return

			case <-e.wakaHeartbeatTicker.C:
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
			if err := e.sendWakaHeartbeat(ctx, program, tracked.Category); err != nil {
				return err
			}
		}
	}

	return nil
}

// Call the wakatime-cli heartbeat command
func (e *EventController) sendWakaHeartbeat(ctx context.Context, program, category string) error {
	cmd := exec.CommandContext(ctx, "wakatime-cli", "--entity", program, "--category", category, "--time", fmt.Sprintf("%d", time.Now().Unix()))
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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
