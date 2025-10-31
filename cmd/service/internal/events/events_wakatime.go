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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

		for {
			select {
			case <-ctx.Done():
				logger.Println("INFO: Stopping heartbeats")
				return
			case <-ticker.C:
				e.sendHeartbeats(ctx, logger, sm)
			}
		}
	}(newCtx)
}

// Send specified heartbeats to WakaTime/Wakapi
func (e *EventController) sendHeartbeats(ctx context.Context, logger *log.Logger, sm *sessions.SessionManager) {
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
}

// Call the wakatime-cli heartbeat command
func (e *EventController) sendWakaTimeHeartbeat(ctx context.Context, logger *log.Logger, program, category, project string) error {
	cliPath := e.Config.WakaTime.CLIPath

	if cliPath == "" {
		return fmt.Errorf("wakatime-cli path not set")
	}

	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		return fmt.Errorf("wakatime-cli not found at path: %s", cliPath)
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
		"--plugin", "timekeep/" + e.version,
	}

	if projectToUse != "" {
		args = append(args, "--project", projectToUse)
	}

	args = append(args,
		"--time", fmt.Sprintf("%f", float64(time.Now().Unix())),
		"--verbose",
	)

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, cliPath, args...)

	cmd.Env = append(os.Environ(),
		"HOME="+os.Getenv("HOME"),
		"PATH="+os.Getenv("PATH"),
		"WAKATIME_HOME="+filepath.Join(os.Getenv("HOME"), ".wakatime"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()

		if exitCode == 112 {
			logger.Printf("INFO: wakatime-cli queued heartbeat (exit 112) in %v", duration)
			if stdout.Len() > 0 {
				logger.Printf("DEBUG: stdout: %s", stdout.String())
			}
			return nil
		}

		if exitCode == 102 {
			logger.Printf("WARNING: wakatime-cli API issue (exit 102) in %v", duration)
			if stderr.Len() > 0 {
				logger.Printf("DEBUG: stderr: %s", stderr.String())
			}
			return nil
		}

		logger.Printf("ERROR: wakatime-cli failed with exit code %d after %v", exitCode, duration)
		logger.Printf("ERROR: stdout: %s", stdout.String())
		logger.Printf("ERROR: stderr: %s", stderr.String())
		return fmt.Errorf("wakatime-cli exited with code %d", exitCode)
	} else if err != nil {
		logger.Printf("ERROR: wakatime-cli failed after %v: %v", duration, err)
		logger.Printf("ERROR: stdout: %s", stdout.String())
		logger.Printf("ERROR: stderr: %s", stderr.String())
		return fmt.Errorf("wakatime-cli execution failed: %v", err)
	}

	return nil
}

// Send heartbeat to user's wakapi server
func (e *EventController) sendWakapiHeartbeat(ctx context.Context, program, category, project string) error {
	if e.Config.Wakapi.Server == "" || e.Config.Wakapi.APIKey == "" {
		return fmt.Errorf("missing config variable")
	}

	apiURL, err := e.validateAndFormatWakapiURL(e.Config.Wakapi.Server)
	if err != nil {
		return fmt.Errorf("invalid Wakapi server URL: %v", err)
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

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(heartbeatData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(e.Config.Wakapi.APIKey)))

	userAgent := e.getUserAgent()
	req.Header.Set("User-Agent", userAgent)

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

// Cancels heartbeat context
func (e *EventController) StopHeartbeats() {
	e.mu.Lock()
	cancel := e.WakaCancel
	e.WakaCancel = nil
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Format URL for Wakapi server address
func (e *EventController) validateAndFormatWakapiURL(server string) (string, error) {
	if server == "" {
		return "", fmt.Errorf("server URL cannot be empty")
	}

	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}

	parsed, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %v", err)
	}

	formatted := parsed.Scheme + "://" + parsed.Host
	if parsed.Path != "" {
		formatted += parsed.Path
	}

	return strings.TrimRight(formatted, "/") + "/api/heartbeat", nil
}

// Create User Agent header for Wakapi request
func (e *EventController) getUserAgent() string {
	app := "Timekeep"
	version := e.version
	os := runtime.GOOS

	switch os {
	case "windows":
		return fmt.Sprintf("%s/%s (Windows NT 10.0; Win64; x64)", app, version)
	case "linux":
		return fmt.Sprintf("%s/%s (X11; Linux x86_64)", app, version)
	case "darwin":
		return fmt.Sprintf("%s/%s (Macintosh; Intel Mac OS X 10_15_7)", app, version)
	default:
		return fmt.Sprintf("%s/%s (%s)", app, version, os)
	}
}
