//go:build linux

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func RunService(name string, isDebug *bool) error {
	service, err := ServiceSetup()
	if err != nil {
		return err
	}
	status, err := service.Manage()
	if err != nil {
		service.logger.Printf("%s: %v", status, err)
		return err
	}

	fmt.Println(status)
	return nil
}

// Main daemon management function
func (s *timekeepService) Manage() (string, error) {
	usage := "Usage: timekeep install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return s.daemon.Install()
		case "remove":
			return s.daemon.Remove()
		case "start":
			return s.daemon.Start()
		case "stop":
			return s.daemon.Stop()
		case "status":
			return s.daemon.Status()
		default:
			return usage, nil
		}
	}

	// Service goroutines here

	go s.listenSocket()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	killSignal := <-interrupt
	s.logger.Printf("Got signal: %v", killSignal)
	close(s.shutdown)
	s.fileCleanup()

	if killSignal == os.Interrupt {
		return "Daemon was interrupted by system signal.", nil
	}

	return "Daemon was killed.", nil
}

func (s *timekeepService) listenSocket() {
	socketName := "/tmp/timekeep.sock"

	listener, err := net.Listen("unix", socketName)
	if err != nil {
		s.logger.Printf("ERROR: Failed to open socket connection")
		return
	}
	defer os.Remove(socketName)
	defer listener.Close()

	s.logger.Printf("INFO: Listening on Unix socket: %s", socketName)

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.logger.Printf("ERROR: Failed to accept connection: %s", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

func (s *timekeepService) startProcessMonitor(programs []string) {
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
		} // skip non-numeric dirs

		commPath := filepath.Join("/proc", name, "comm")
		b, err := os.ReadFile(commPath)
		if err != nil {
			// process may have exited; ignore ENOENT
			if !errors.Is(err, fs.ErrNotExist) {
				s.logger.Printf("read %s: %v", commPath, err)
			}
			continue
		}
		comm := strings.TrimSpace(string(b))
		// use comm...
	}
}

func getLogPath() (string, error) {
	logDir := "/tmp/Timekeep/logs"
	return filepath.Join(logDir, "timekeep.log"), nil
}
