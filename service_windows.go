//go:build windows
// +build windows

package main

import (
	"log"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

/* Interface to be implemented to build a Windows Service
type Handler interface {
	//Will be called by package code at start of service, and will exit once Execute is complete
	//Read service change requests from 'r', and keep service control manager up to date about service state by writing into 's'
	//Args contains service name followed by argument strings
	//Can provide service exit code in exitCode return parameter, also indicate if exit code, if any, is service specific or not using svcSpecificEC
	Execute(args []string, r <-chan ChangeRequest, s chan<- Status) (svcSpecificEC bool, exitCode uint32)
}*/

const serviceName = "ProcessTracker"

// Service context
type processTrackerService struct{}

func (s *processTrackerService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {

	//Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	tick := time.Tick(30 * time.Second)

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	//Service mainloop
loop:
	for {
		//Receive channel signals
		select {

		//Tick signal
		case <-tick:
			log.Printf("Tick Handled...!")

		//'r' receive only channel, SCM signals
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: //Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: //Service needs to be stopped or shutdown
				log.Printf("Shutting service...!")
				break loop
			case svc.Pause: //Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: //Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				log.Printf("Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

func RunService(name string, isDebug bool) {
	if isDebug {
		err := debug.Run(name, &processTrackerService{})
		if err != nil {
			log.Fatalln("Error running service in debug mode.")
		}
	} else {
		err := svc.Run(name, &processTrackerService{})
		if err != nil {
			log.Fatalln("Error running service in Service Control mode.")
		}
	}
}
