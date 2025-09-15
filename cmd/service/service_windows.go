//go:build windows

package main

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

//go:embed monitor.ps1
var monitorScript string

func RunService(name string, isDebug *bool) error {
	if *isDebug {
		service, err := TestServiceSetup()
		if err != nil {
			return err
		}
		return debug.Run(name, service)
	} else {
		service, err := ServiceSetup()
		if err != nil {
			return err
		}
		return svc.Run(name, service)
	}
}

// Service execute method for Windows Handler interface
func (s *timekeepService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	s.logger.Println("INFO: Service Execute function entered.")

	if s.logFile != nil {
		err := s.logFile.Sync()
		if err != nil {
			s.logger.Printf("ERROR: Failed to sync log file: %v", err)
		}
	}

	// Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	programs, err := s.prRepo.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Printf("ERROR: Failed to get programs: %s", err)
		return false, 1
	}
	if len(programs) > 0 {
		s.startProcessMonitor(programs)
	}

	go s.listenPipe()

	// Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: // Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: // Service needs to be stopped or shutdown
				close(s.shutdown)
				s.fileCleanup()
				break loop
			case svc.Pause: // Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: // Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				s.logger.Printf("ERROR: Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 0
}

// Opens a Windows named pipe connection, to listen for commands
func (s *timekeepService) listenPipe() {
	pipeName := "\\\\.\\pipe\\Timekeep"

	pipe, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		s.logger.Printf("ERROR: Failed to create pipe: %s", err)
		return
	}
	defer pipe.Close()

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := pipe.Accept()
			if err != nil {
				s.logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

// Runs the powershell WMI script, to monitor process events
func (s *timekeepService) startProcessMonitor(programs []string) {
	programList := strings.Join(programs, ",")

	scriptTempDir := filepath.Join("C:\\", "ProgramData", "TimeKeep", "scripts_temp")

	if err := os.MkdirAll(scriptTempDir, 0o755); err != nil {
		s.logger.Printf("ERROR: Failed to create PowerShell script temp directory '%s': %s", scriptTempDir, err)
		return
	}

	tempFile, err := os.CreateTemp(scriptTempDir, "monitor*.ps1")
	if err != nil {
		s.logger.Printf("ERROR: Failed to create temp script file in '%s': %s", scriptTempDir, err)
		return
	}

	defer tempFile.Close()

	if _, err := tempFile.WriteString(monitorScript); err != nil {
		s.logger.Printf("ERROR: Failed to write script: %s", err)
		return
	}

	if err := tempFile.Sync(); err != nil {
		s.logger.Printf("ERROR: Failed to sync temp script file to disk: %s", err)
		return
	}

	tempFile.Close()

	time.Sleep(100 * time.Millisecond) // Pause to allow tempfile to finish writing before it attempts to execute

	args := []string{"-ExecutionPolicy", "Bypass", "-File", tempFile.Name(), "-Programs", programList}
	cmd := exec.Command("powershell", args...)
	s.psProcess = cmd

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		s.logger.Printf("ERROR: Failed to start PowerShell monitor: %s", err)
		s.psProcess = nil
		if stderr.Len() > 0 {
			s.logger.Printf("INFO: PowerShell stderr (on Start() failure): %s", stderr.String())
		}
	}

	// Goroutine to wait for the PowerShell process to exit and log its stderr/stdout
	go func() {
		defer os.Remove(tempFile.Name())

		err := cmd.Wait()
		if err != nil {
			s.logger.Printf("ERROR: PowerShell monitor process exited with error: %s", err)
		} else {
			s.logger.Println("INFO: PowerShell monitor process exited successfully.")
		}

		if stderr.Len() > 0 {
			s.logger.Printf("PowerShell stderr output: %s", stderr.String())
		} else {
			s.logger.Println("INFO: No PowerShell stderr output.")
		}
	}()
}

// Get path for logging file
func getLogPath() (string, error) {
	logDir := `C:\ProgramData\TimeKeep\logs`
	return filepath.Join(logDir, "timekeep.log"), nil
}
