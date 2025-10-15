//go:build windows

package events

import (
	"bytes"
	"context"
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

//go:embed monitor.ps1
var monitorScript string

//go:embed premonitor.ps1
var premonitorScript string

// Main process monitoring function for Windows version
func (e *EventController) MonitorProcesses(ctx context.Context, logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	e.startProcessMonitor(ctx, logger, programs)
}

// Runs the powershell WMI script, to monitor process events
func (e *EventController) startProcessMonitor(ctx context.Context, logger *log.Logger, programs []string) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		logger.Println("WARNING: Context already cancelled, not starting monitor")
		return
	default:
	}

	programList := strings.Join(programs, ",")

	scriptTempDir := filepath.Join("C:\\", "ProgramData", "TimeKeep", "scripts_temp")

	if err := os.MkdirAll(scriptTempDir, 0o755); err != nil {
		logger.Printf("ERROR: Failed to create PowerShell script temp directory '%s': %s", scriptTempDir, err)
		return
	}

	tempFile, err := os.CreateTemp(scriptTempDir, "monitor*.ps1")
	if err != nil {
		logger.Printf("ERROR: Failed to create temp script file in '%s': %s", scriptTempDir, err)
		return
	}

	defer tempFile.Close()

	if _, err := tempFile.WriteString(monitorScript); err != nil {
		logger.Printf("ERROR: Failed to write script: %s", err)
		return
	}

	if err := tempFile.Sync(); err != nil {
		logger.Printf("ERROR: Failed to sync temp script file to disk: %s", err)
		return
	}

	tempFile.Close()

	time.Sleep(100 * time.Millisecond) // Pause to allow tempfile to finish writing before it attempts to execute

	args := []string{"-ExecutionPolicy", "Bypass", "-File", tempFile.Name(), "-Programs", programList}
	cmd := exec.CommandContext(ctx, "powershell", args...)
	e.PsProcess = cmd

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logger.Println("INFO: Executing monitor script")
	if err := cmd.Start(); err != nil {
		logger.Printf("ERROR: Failed to start PowerShell monitor: %s", err)
		e.PsProcess = nil
		if stderr.Len() > 0 {
			logger.Printf("INFO: PowerShell stderr (on Start() failure): %s", stderr.String())
		}
	}

	// Goroutine to wait for the PowerShell process to exit and log its stderr/stdout
	go func() {
		defer os.Remove(tempFile.Name())

		err := cmd.Wait()

		select {
		case <-ctx.Done():
			logger.Println("INFO: Powershell monitor stopped due to context cancellation")
			return
		default:
		}

		if err != nil {
			logger.Printf("ERROR: PowerShell monitor process exited with error: %s", err)
		} else {
			logger.Println("INFO: PowerShell monitor process exited successfully.")
		}

		if stderr.Len() > 0 {
			logger.Printf("PowerShell stderr output: %s", stderr.String())
		} else {
			logger.Println("INFO: No PowerShell stderr output.")
		}
	}()
}

// Stops the WMI powershell script
func (e *EventController) StopProcessMonitor() {
	if e.PsProcess != nil {
		_ = e.PsProcess.Process.Kill()
		e.PsProcess = nil
	}
}

// Runs the pre-monitoring script, gathering PIDs for tracked programs that are already running on service start
func (e *EventController) StartPreMonitor(ctx context.Context, logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository, programs []string) {
	select {
	case <-ctx.Done():
		logger.Println("WARNING: Context already cancelled, not starting pre-monitor")
		return
	default:
	}

	programList := strings.Join(programs, ",")

	scriptTempDir := filepath.Join("C:\\", "ProgramData", "TimeKeep", "scripts_temp")

	if err := os.MkdirAll(scriptTempDir, 0o755); err != nil {
		logger.Printf("ERROR: Failed to create PowerShell script temp directory '%s': %s", scriptTempDir, err)
		return
	}

	tempFile, err := os.CreateTemp(scriptTempDir, "premonitor*.ps1")
	if err != nil {
		logger.Printf("ERROR: Failed to create temp script file in '%s': %s", scriptTempDir, err)
		return
	}

	defer tempFile.Close()

	if _, err := tempFile.WriteString(premonitorScript); err != nil {
		logger.Printf("ERROR: Failed to write script: %s", err)
		return
	}

	if err := tempFile.Sync(); err != nil {
		logger.Printf("ERROR: Failed to sync temp script file to disk: %s", err)
		return
	}

	tempFile.Close()

	time.Sleep(100 * time.Millisecond)

	args := []string{"-ExecutionPolicy", "Bypass", "-File", tempFile.Name(), "-Programs", programList}
	cmd := exec.CommandContext(ctx, "powershell", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logger.Println("INFO: Executing pre-monitor script")
	if err := cmd.Start(); err != nil {
		logger.Printf("ERROR: Failed to start PowerShell script: %s", err)
		if stderr.Len() > 0 {
			logger.Printf("INFO: PowerShell stderr (on Start() failure): %s", stderr.String())
		}
	}

	go func() {
		defer os.Remove(tempFile.Name())

		err := cmd.Wait()

		select {
		case <-ctx.Done():
			logger.Println("INFO: Powershell pre-monitor stopped due to context cancellation")
			return
		default:
		}

		if err != nil {
			logger.Printf("ERROR: PowerShell pre-monitor process exited with error: %s", err)
		} else {
			logger.Println("INFO: PowerShell pre-monitor process exited successfully.")
		}

		if stderr.Len() > 0 {
			logger.Printf("PowerShell stderr output: %s", stderr.String())
		} else {
			logger.Println("INFO: No PowerShell stderr output.")
		}
	}()
}
