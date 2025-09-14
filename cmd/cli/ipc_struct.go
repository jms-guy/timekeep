package main

type (
	realServiceCommander struct{}
	testServiceCommander struct{}
)

type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

type ServiceCommander interface {
	WriteToService() error
}

func (r *testServiceCommander) WriteToService() error {
	return nil
}
