package main

import (
	"context"
)

type (
	realCommandExecutor struct{}
	testCommandExecutor struct{}
)

type CommandExecutor interface {
	RunCommand(ctx context.Context, name string, args ...string) (stdout string, err error)
}

func (t *testCommandExecutor) RunCommand(ctx context.Context, name string, args ...string) (stdout string, err error) {
	returnStr := "SERVICE_NAME: Timekeep\nTYPE               : 10  WIN32_OWN_PROCESS\nSTATE              : 4  RUNNING\n(STOPPABLE, PAUSABLE, ACCEPTS_SHUTDOWN)\nWIN32_EXIT_CODE    : 0  (0x0)\nSERVICE_EXIT_CODE  : 0  (0x0)\nCHECKPOINT         : 0x0\nWAIT_HINT          : 0x0"
	return returnStr, nil
}
