//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/Microsoft/go-winio"
	"github.com/jms-guy/timekeep/internal/database"
	"github.com/jms-guy/timekeep/sql"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const serviceName = "TimeKeep"

// Service context
type timekeepService struct {
	db       *database.Queries
	logger   *log.Logger
	shutdown chan struct{}
}

type Command struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data,omitempty"`
}

func (s *timekeepService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	s.shutdown = make(chan struct{})

	//Signals that service can accept from SCM(Service Control Manager)
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go s.listenPipe()

	//Service mainloop, handles only SCM signals
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate: //Check current status of service
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown: //Service needs to be stopped or shutdown
				close(s.shutdown)
				break loop
			case svc.Pause: //Service needs to be paused, without shutdown
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue: //Resume paused execution state of service
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				s.logger.Printf("Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

// Opens a Windows named pipe connection, to listen for commands
func (s *timekeepService) listenPipe() {
	pipeName := `\\.\pipe\TimeKeep`

	for {
		select {
		case <-s.shutdown:
			return
		default:
			pipe, err := winio.ListenPipe(pipeName, nil)
			if err != nil {
				s.logger.Printf("Failed to create pipe: %w", err)
				continue
			}

			conn, err := pipe.Accept()
			if err != nil {
				s.logger.Printf("Failed to accept connection: %w", err)
				continue
			}

			go s.handlePipeConnection(conn)
		}
	}
}

// Handles service commands read from pipe connection
func (s *timekeepService) handlePipeConnection(conn net.Conn) {
	defer conn.Close()

	var cmd Command
	decoder := json.NewDecoder(conn)

	if err := decoder.Decode(&cmd); err != nil {
		s.logger.Printf("Failed to decode command: %w", err)
		return
	}

	switch cmd.Action {
	case "add_program":
		s.addProgram(cmd.Data["name"].(string))
	case "remove_program":
		s.removeProgram(cmd.Data["name"].(string))
	case "get_stats":
		stats := s.getStats()
		json.NewEncoder(conn).Encode(stats)
	}
}

// Creates new service instance
func NewTimekeepService() (*timekeepService, error) {
	db, err := sql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "TimeKeep: ", log.LstdFlags)

	return &timekeepService{
		db:     db,
		logger: logger,
	}, nil
}

func RunService(name string, isDebug bool) error {
	service, err := NewTimekeepService()
	if err != nil {
		return err
	}
	if isDebug {
		return debug.Run(name, service)
	} else {
		return svc.Run(name, service)
	}
}
