package events

import (
	"time"

	"github.com/jms-guy/timekeep/cmd/service/internal/sessions"
)

// Start WakaTime heartbeat ticker
func (e *EventController) StartHeartbeats(s *sessions.SessionManager) {
	e.wakaHeartbeatTicker = time.NewTicker(1 * time.Minute)

	go func() {
		for range e.wakaHeartbeatTicker.C {
			e.sendHeartbeats(s)
		}
	}()
}

// Send specified heartbeats to WakaTime
func (e *EventController) sendHeartbeats(s *sessions.SessionManager) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for program, tracked := range s.Programs {
		if len(tracked.PIDs) > 0 {
			e.sendWakaHeartbeat()
		}
	}
}

// Stops WakaTime heartbeat ticker after disabling integration
func (e *EventController) StopHeartbeats() {
	e.wakaHeartbeatTicker.Stop()
}
