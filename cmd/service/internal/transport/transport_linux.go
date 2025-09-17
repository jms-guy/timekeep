//go:build linux

package transport

import (
	"log"
	"net"
	"os"

	"github.com/jms-guy/timekeep/cmd/service/internal/events"
	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (t *Transporter) Listen(logger *log.Logger, eventCtrl *events.EventController, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	socketName := "/tmp/timekeep.sock"

	listener, err := net.Listen("unix", socketName)
	if err != nil {
		logger.Printf("ERROR: Failed to open socket connection")
		return
	}
	defer os.Remove(socketName)
	defer listener.Close()

	logger.Printf("INFO: Listening on Unix socket: %s", socketName)

	for {
		select {
		case <-t.Shutdown:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go eventCtrl.HandleConnection(logger, s, pr, a, h, conn)
		}
	}
}
