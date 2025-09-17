package daemons

type DaemonManager interface {
	Install() (string, error)
	Remove() (string, error)
	Start() (string, error)
	Stop() (string, error)
	Status() (string, error)
}
