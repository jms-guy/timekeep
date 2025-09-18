//go:build !windows && !linux

package events

import (
	"log"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
	"github.com/jms-guy/timekeep/internal/repository"
)

func (e *EventController) MonitorProcesses(logger *log.Logger, s *sessions.SessionManager, pr repository.ProgramRepository, a repository.ActiveRepository, h repository.ActiveRepository, programs []string) {
	return
}

func (e *EventController) startProcessMonitor(logger *log.Logger, programs []string) {
	return
}

func (e *EventController) StopProcessMonitor() {
	return
}
