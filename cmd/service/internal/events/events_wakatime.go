package events

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
)

// Start WakaTime/Wakapi heartbeat ticker
func (e *EventController) StartHeartbeats(parent context.Context, logger *log.Logger, sm *sessions.SessionManager) {
	newCtx, newCancel := context.WithCancel(parent)

	e.mu.Lock()
	oldCancel := e.WakaCancel
	e.WakaCancel = newCancel
	e.mu.Unlock()

	if oldCancel != nil {
		oldCancel()
	}

	if e.Config.Wakapi.Enabled && e.Client == nil {
		logger.Println("INFO: Initializing Wakapi Http client")
		e.Client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: false,
				MaxIdleConns:      10,
				IdleConnTimeout:   90 * time.Second,
			},
		}
	}

	logger.Println("INFO: Starting heartbeats")

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		errorCount := 0
		for {
			select {
			case <-ctx.Done():
				logger.Println("INFO: Stopping heartbeats")
				return
			case <-ticker.C:
				if errorCount >= 5 {
					logger.Println("ERROR: Heartbeats failed 5 times consecutively, stopping")
					return
				}
				if err := e.sendHeartbeats(ctx, logger, sm); err != nil {
					errorCount++
					continue
				}
				errorCount = 0
			}
		}
	}(newCtx)
}

// Send specified heartbeats to WakaTime/Wakapi
func (e *EventController) sendHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) error {
	type item struct{ program, category, project string }
	items := []item{}
	var err error

	sm.Mu.Lock()
	for p, t := range sm.Programs {
		if len(t.PIDs) > 0 && t.Category != "" {
			items = append(items, item{p, t.Category, t.Project})
		}
	}
	sm.Mu.Unlock()

	for _, it := range items {
		if e.Config.WakaTime.Enabled {
			if err = e.sendWakaTimeHeartbeat(ctx, logger, it.program, it.category, it.project); err != nil {
				logger.Printf("ERROR: Failed to send WakaTime heartbeat: %s", err)
			} else {
				logger.Printf("INFO: WakaTime heartbeat sent for %s, category %s", it.program, it.category)
			}
		}

		if e.Config.Wakapi.Enabled {
			if err := e.sendWakapiHeartbeat(ctx, it.program, it.category, it.project); err != nil {
				logger.Printf("ERROR: Failed to send Wakapi heartbeat: %s", err)
			} else {
				logger.Printf("INFO: Wakapi heartbeat send for %s, category %s", it.program, it.category)
			}
		}
	}

	return err
}

// Call the wakatime-cli heartbeat command
func (e *EventController) sendWakaTimeHeartbeat(ctx context.Context, logger *log.Logger, program, category, project string) error {
	cliPath := e.Config.WakaTime.CLIPath

	if cliPath == "" {
		return fmt.Errorf("wakatime-cli path not set")
	}

	projectToUse := e.Config.WakaTime.GlobalProject
	if project != "" {
		projectToUse = project
	}

	args := []string{
		"--key", e.Config.WakaTime.APIKey,
		"--entity", program,
		"--entity-type", "app",
		"--category", category,
		"--alternate-project", projectToUse,
		"--time", fmt.Sprintf("%d", time.Now().Unix()),
		"--verbose",
		"--write",
	}

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

// Send heartbeat to user's wakapi instance
func (e *EventController) sendWakapiHeartbeat(ctx context.Context, program, category, project string) error {
	if e.Config.Wakapi.Server == "" || e.Config.Wakapi.APIKey == "" {
		return fmt.Errorf("missing config variable")
	}

	projectToUse := e.Config.Wakapi.GlobalProject
	if project != "" {
		projectToUse = project
	}

	type Heartbeat struct {
		Entity   string `json:"entity"`
		Type     string `json:"type"`
		Category string `json:"category"`
		Project  string `json:"project"`
		Time     int64  `json:"time"`
		IsWrite  bool   `json:"is_write"`
	}

	heartbeat := Heartbeat{
		Entity:   program,
		Type:     "app",
		Category: category,
		Project:  projectToUse,
		Time:     time.Now().Unix(),
		IsWrite:  false,
	}

	heartbeatData, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(e.Config.Wakapi.Server, "/")+"/heartbeat", bytes.NewBuffer(heartbeatData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(e.Config.Wakapi.APIKey)))

	resp, err := e.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wakapi: status %d: %s", resp.StatusCode, bytes.TrimSpace(b))
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
