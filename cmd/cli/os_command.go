package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Run os-specific commands
func (r *realCommandExecutor) RunCommand(ctx context.Context, name string, args ...string) (stdout string, err error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdoutBuf.String(), fmt.Errorf("error running external command: %w: %s", exitErr, stderrBuf.String())
		}
		return stdoutBuf.String(), err
	}

	return stdoutBuf.String(), nil
}
