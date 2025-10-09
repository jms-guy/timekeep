package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
)

// Start WakaTime heartbeat ticker
func (e *EventController) StartHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) {
	e.heartbeatMu.Lock()
	e.wakaHeartbeatTicker = time.NewTicker(1 * time.Minute)
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
	heartbeat := map[string]interface{}{
		"entity":   program,
		"type":     "app",
		"category": category,
		"time":     float64(time.Now().Unix()),
	}

	body, _ := json.Marshal([]map[string]interface{}{heartbeat})

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.wakatime.com/api/v1/users/current/heartbeats.bulk",
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+e.Config.WakaTime.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wakatime API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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
