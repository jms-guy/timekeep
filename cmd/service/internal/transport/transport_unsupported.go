//go:build !windows && !linux

package transport

import (
	"log"

	"github.com/jms-guy/timekeep/cmd/service/internal/events"
	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (t *Transporter) Listen(logger *log.Logger, eventCtrl *events.EventController, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.HistoryRepository) {
	return
}
