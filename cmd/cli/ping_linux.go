//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// Gets current service state for user
func (s *CLIService) StatusService() error {
	cmd := exec.Command("systemctl", "is-active", "timekeep.service")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("service not running: %v", err)
	}

	status := strings.TrimSpace(string(output))
	if status != "active" {
		return fmt.Errorf("service is not active; Status: %s", status)
	}

	fmt.Printf("Service Status: %s\n", status)

	return nil
}
