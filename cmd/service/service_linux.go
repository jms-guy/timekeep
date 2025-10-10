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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	programs, err := s.prRepo.GetAllPrograms(ctx)
	if err != nil {
		return "ERROR: Failed to get programs", err
	}
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
			s.sessions.EnsureProgram(program.Name, category, project)

			toTrack = append(toTrack, program.Name)
		}

		go s.eventCtrl.MonitorProcesses(ctx, s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, toTrack)
	}

	if s.eventCtrl.Config.WakaTime.Enabled {
		s.eventCtrl.StartHeartbeats(ctx, s.logger.Logger, s.sessions)
	}

	go s.transport.Listen(ctx, s.logger.Logger, s.eventCtrl, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

	<-ctx.Done()

	s.logger.Logger.Println("INFO: Received shutdown signal")
	s.closeService(ctx)

	return "INFO: Daemon stopped.", nil
}
