//go:build windows

package main

import (
	"context"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

// Windows specific service management functions

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

	serviceCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	programs, err := s.prRepo.GetAllPrograms(context.Background())
	if err != nil {
		s.logger.Logger.Printf("ERROR: Failed to get programs: %s", err)
		status <- svc.Status{State: svc.Stopped}
		return false, 1
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
			s.sessions.Mu.Lock()
			s.sessions.EnsureProgram(program.Name, category, project)
			s.sessions.Mu.Unlock()

			toTrack = append(toTrack, program.Name)
		}

		s.eventCtrl.StartPreMonitor(s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, toTrack)
		s.eventCtrl.StartMonitor(serviceCtx, s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo, toTrack)
	}

	if s.eventCtrl.Config.WakaTime.Enabled || s.eventCtrl.Config.Wakapi.Enabled {
		s.eventCtrl.StartHeartbeats(serviceCtx, s.logger.Logger, s.sessions)
	}

	go s.transport.Listen(serviceCtx, s.logger.Logger, s.eventCtrl, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: // Check current status of service
				status <- c.CurrentStatus

			case svc.Stop, svc.Shutdown: // Service needs to be stopped or shutdown
				status <- svc.Status{State: svc.StopPending}
				s.logger.Logger.Println("INFO: Received stop signal")
				s.closeService(s.logger.Logger)
				s.eventCtrl.MonCancel()
				s.eventCtrl.WakaCancel()
				cancel()
				break loop

			case svc.Pause: // Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				s.logger.Logger.Println("INFO: Pausing service")
				if s.eventCtrl.Config.WakaTime.Enabled {
					s.eventCtrl.StopHeartbeats()
				}
				s.eventCtrl.StopProcessMonitor()

			case svc.Continue: // Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				s.logger.Logger.Println("INFO: Resuming service")
				s.eventCtrl.RefreshProcessMonitor(serviceCtx, s.logger.Logger, s.sessions, s.prRepo, s.asRepo, s.hsRepo)

			default:
				s.logger.Logger.Printf("ERROR: Unexpected service control request #%d", c)
			}
		}
	}

	return false, 0
}
