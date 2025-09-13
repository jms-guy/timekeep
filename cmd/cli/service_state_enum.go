package main

type ServiceState int

const (
	Ignore ServiceState = iota
	Stopped
	Start_Pending
	Stop_Pending
	Running
	Continue_Pending
	Pause_Pending
	Paused
)

var stateName = map[ServiceState]string{
	Stopped:          "Stopped",
	Start_Pending:    "Start Pending",
	Stop_Pending:     "Stop Pending",
	Running:          "Running",
	Continue_Pending: "Continue Pending",
	Pause_Pending:    "Pause Pending",
	Paused:           "Paused",
}
