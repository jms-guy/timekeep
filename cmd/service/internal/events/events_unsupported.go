//go:build !windows && !linux

package events

import "log"

func (e *EventController) StartProcessMonitor(logger *log.Logger, programs []string) {
	return
}

func (e *EventController) StopProcessMonitor() {
	return
}
