//go:build !windows && !linux

package main

import (
	"log"
)

func RunService(name string, isDebug bool) {
	log.Fatal("Unsupported platform")
}
