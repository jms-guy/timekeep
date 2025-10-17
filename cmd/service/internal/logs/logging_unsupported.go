//go:build !windows && !linux

package logs

import (
	"log"
	"os"
)

func getLogPath() (string, error) {
	return "", nil
}

func CreateLogger(logPath string) (*log.Logger, *os.File, error) {
	return nil, nil, nil
}
