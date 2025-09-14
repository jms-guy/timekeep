//go:build !windows && !linux

package main

import (
	"log"
)

func RunService(name string, isDebug *bool) error {
	log.Fatal("Unsupported platform")
	return nil
}
