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

	programs, err := s.prRepo.GetAllProgramNames(context.Background())
	if err != nil {
		return "ERROR: Failed to get programs", err
	}
	if len(programs) > 0 {
		for _, program := range programs {
			s.sessions.EnsureProgram(program)
		}
		go s.eventCtrl.MonitorProcesses(s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, programs)
	}

	go s.transport.Listen(s.logger.Logger, s.eventCtrl, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	killSignal := <-interrupt
	s.logger.Logger.Printf("Got signal: %v", killSignal)
	s.closeService()

	if killSignal == os.Interrupt {
		return "INFO: Daemon was interrupted by system signal.", nil
	}

	return "INFO: Daemon was killed.", nil
}
