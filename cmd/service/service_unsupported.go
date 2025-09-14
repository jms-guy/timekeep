//go:build !windows && !linux

package main

import (
	"log"
)

func getLogPath() (string, error) {
	return "", nil
}

func RunService(name string, isDebug *bool) error {
	log.Fatal("Unsupported platform")
	return nil
}
