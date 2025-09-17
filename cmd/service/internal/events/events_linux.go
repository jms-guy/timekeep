//go:build linux

package events

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (e *EventController) StartProcessMonitor(logger *log.Logger, programs []string) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		s.logger.Printf("ERROR: Couldn't read /proc: %s", err)
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		pid, err := strconv.Atoi(name)
		if err != nil {
			continue
		} 

		commPath := filepath.Join("/proc", name, "comm")
		b, err := os.ReadFile(commPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				s.logger.Printf("read %s: %v", commPath, err)
			}
			continue
		}
		comm := strings.TrimSpace(string(b))
		
		if comm
	}
}

func (e *EventController) StopProcessMonitor() {
	return
}