//go:build !windows && !linux

package sql

import "log"

func getDatabasePath(logger *log.Logger) (string, error) {
	return "", nil
}
