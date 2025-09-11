//go:build windows

package main

import (
	"encoding/json"
	"fmt"

	"github.com/Microsoft/go-winio"
)

// Connects to named pipe opened by main service, to communicate an action to the service
func WriteToService() error {
	pipeName := "\\\\.\\pipe\\Timekeep"

	msg := Command{
		Action: "refresh",
	}

	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to service pipe: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to write JSON to pipe: %v", err)
	}

	return nil
}
