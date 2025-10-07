//go:build windows

package main

import (
	"context"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

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
	s.logger.Logger.Println("INFO: Service Execute function entered.")

	if s.logger.LogFile != nil {
		err := s.logger.LogFile.Sync()
		if err != nil {
			s.logger.Logger.Printf("ERROR: Failed to sync log file: %v", err)
		}
	}

	// Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	programs, err := s.prRepo.GetAllProgramNames(context.Background())
	if err != nil {
		s.logger.Logger.Printf("ERROR: Failed to get programs: %s", err)
		return false, 1
	}
	if len(programs) > 0 {
		for _, program := range programs {
			s.sessions.EnsureProgram(program)
		}
		s.eventCtrl.MonitorProcesses(s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, programs)
	}

	if s.eventCtrl.Config.WakaTime.Enabled {
		s.eventCtrl.StartHeartbeats(s.sessions)
	}

	go s.transport.Listen(s.logger.Logger, s.eventCtrl, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

	// Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: // Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: // Service needs to be stopped or shutdown
				s.closeService()
				break loop
			case svc.Pause: // Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: // Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				s.logger.Logger.Printf("ERROR: Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 0
}
