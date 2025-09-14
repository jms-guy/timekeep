//go:build linux

package main

import (
	"context"
)

// Run os-specific commands
func (r *realCommandExecutor) RunCommand(ctx context.Context, name string, args ...string) (stdout string, err error) {
	return "", nil
}
