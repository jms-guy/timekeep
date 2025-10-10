//go:build linux

package transport

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jms-guy/timekeep/cmd/service/internal/events"
	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (t *Transporter) Listen(ctx context.Context, logger *log.Logger, eventCtrl *events.EventController, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	socketDir := "/var/run/timekeep"
	socketName := socketDir + "/timekeep.sock"

	if err := os.MkdirAll(socketDir, 0o755); err != nil {
		logger.Printf("ERROR: Failed to create socket directory: %v", err)
		return
	}

	os.Remove(socketName)

	listener, err := net.Listen("unix", socketName)
	if err != nil {
		logger.Printf("ERROR: Failed to open socket connection")
		return
	}

	if err := os.Chmod(socketName, 0o666); err != nil {
		logger.Printf("WARNING: Could not set socket permissions: %v", err)
	}

	defer os.Remove(socketName)
	defer listener.Close()

	logger.Printf("INFO: Listening on Unix socket: %s", socketName)

	for {
		select {
		case <-ctx.Done():
			logger.Println("INFO: Closing socket connection")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go eventCtrl.HandleConnection(ctx, logger, s, pr, a, h, conn)
		}
	}
}
