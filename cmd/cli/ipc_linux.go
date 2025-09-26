//go:build linux

package main

import (
	"encoding/json"
	"fmt"
	"net"
)

func (r *realServiceCommander) WriteToService() error {
	socketName := "/tmp/timekeep.sock"

	msg := Command{
		Action: "refresh",
	}

	conn, err := net.Dial("unix", socketName)
	if err != nil {
		return fmt.Errorf("failed to connect to socket: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to write JSON to socket: %v", err)
	}

	return nil
}
