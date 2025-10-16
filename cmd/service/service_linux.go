//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func RunService(name string, isDebug *bool) error {
	service, err := ServiceSetup()
	if err != nil {
		return err
	}
	status, err := service.Manage()
	if err != nil {
		service.logger.Logger.Printf("%s: %v", status, err)
		return err
	}

	fmt.Println(status)
	return nil
}

// Main daemon management function
func (s *timekeepService) Manage() (string, error) {
	logger := s.logger.Logger

	logger.Println("INFO: Starting Manage function")
	usage := "Usage: timekeep install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return s.daemon.Install()
		case "remove":
			return s.daemon.Remove()
		case "start":
			return s.daemon.Start()
		case "stop":
			return s.daemon.Stop()
		case "status":
			return s.daemon.Status()
		default:
			return usage, nil
		}
	}

	serviceCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runCtx, runCancel := context.WithCancel(serviceCtx)
	s.eventCtrl.RunCtx = runCtx
	s.eventCtrl.Cancel = runCancel

	logger.Printf("DEBUG: Getting initial programs")
	programs, err := s.prRepo.GetAllPrograms(context.Background())
	if err != nil {
		return "ERROR: Failed to get programs", err
	}
	logger.Printf("DEBUG: Have %d programs", len(programs))
	if len(programs) > 0 {
		toTrack := []string{}
		for _, program := range programs {
			category := ""
			if program.Category.Valid {
				category = program.Category.String
			}
			project := ""
			if program.Project.Valid {
				project = program.Project.String
			}
			logger.Printf("DEBUG: Tracking %s", program.Name)
			s.sessions.EnsureProgram(program.Name, category, project)

			toTrack = append(toTrack, program.Name)
		}

		logger.Printf("DEBUG: Entering main Monitor function")
		go s.eventCtrl.MonitorProcesses(runCtx, s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, toTrack)
	}

	logger.Printf("DEBUG: Starting heartbeats")
	if s.eventCtrl.Config.WakaTime.Enabled {
		s.eventCtrl.StartHeartbeats(runCtx, s.logger.Logger, s.sessions)
	}

	go s.transport.Listen(serviceCtx, s.logger.Logger, s.eventCtrl, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

	<-serviceCtx.Done()

	s.logger.Logger.Println("INFO: Received shutdown signal")
	s.closeService(s.logger.Logger)

	return "INFO: Daemon stopped.", nil
}
