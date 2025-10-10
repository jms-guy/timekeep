//go:build windows

package transport

import (
	"context"
	"log"

	"github.com/Microsoft/go-winio"
	"github.com/jms-guy/timekeep/cmd/service/internal/events"
	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

// Opens a Windows named pipe connection, to listen for commands
func (t *Transporter) Listen(ctx context.Context, logger *log.Logger, eventCtrl *events.EventController, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	pipeName := "\\\\.\\pipe\\Timekeep"

	pipe, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		logger.Printf("ERROR: Failed to create pipe: %s", err)
		return
	}
	defer pipe.Close()

	for {
		select {
		case <-ctx.Done():
			logger.Println("INFO: Stopping pipe listener")
			return
		default:
			conn, err := pipe.Accept()
			if err != nil {
				logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go eventCtrl.HandleConnection(ctx, logger, s, pr, a, h, conn)
		}
	}
}
