//go:build windows

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type ServiceState int

const (
	Ignore ServiceState = iota
	Stopped
	Start_Pending
	Stop_Pending
	Running
	Continue_Pending
	Pause_Pending
	Paused
)

var stateName = map[ServiceState]string{
	Stopped:          "Stopped",
	Start_Pending:    "Start Pending",
	Stop_Pending:     "Stop Pending",
	Running:          "Running",
	Continue_Pending: "Continue Pending",
	Pause_Pending:    "Pause Pending",
	Paused:           "Paused",
}

// Gets current service state for user
func (s *CLIService) StatusService() error {
	stdoutResult, err := s.CmdExe.RunCommand(context.Background(), "sc.exe", "query", "Timekeep")
	if err != nil {
		return err
	}

	stdoutLines := strings.Split(stdoutResult, "\n")

	stateStr := ""
	for _, line := range stdoutLines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "STATE") {
			stateStr = line
			break
		}
	}
	if stateStr == "" {
		return fmt.Errorf("missing service state value")
	}

	parts := strings.Fields(stateStr)
	if len(parts) < 3 {
		return fmt.Errorf("malformed state line: %s", stateStr)
	}

	stateValStr := parts[2]
	stateNum, err := strconv.Atoi(stateValStr)
	if err != nil {
		return fmt.Errorf("error converting state number '%s' to integer: %w", stateValStr, err)
	}

	if state, ok := stateName[ServiceState(stateNum)]; ok {
		fmt.Printf("Service status: %s\n", state)
	} else {
		fmt.Printf("Service status: Unknown state (%d)\n", stateNum)
	}

	return nil
}
