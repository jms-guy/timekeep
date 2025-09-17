//go:build linux

package daemons

import "github.com/takama/daemon"

type linuxDaemon struct {
	d daemon.Daemon
}

func NewDaemonManager() (DaemonManager, error) {
	d, err := daemon.New("timekeep", "Timekeep Process Tracker", daemon.SystemDaemon)
	if err != nil {
		return nil, err
	}
	return &linuxDaemon{d: d}, nil
}

func (l *linuxDaemon) Install() (string, error) { return l.d.Install() }
func (l *linuxDaemon) Remove() (string, error)  { return l.d.Remove() }
func (l *linuxDaemon) Start() (string, error)   { return l.d.Start() }
func (l *linuxDaemon) Stop() (string, error)    { return l.d.Stop() }
func (l *linuxDaemon) Status() (string, error)  { return l.d.Status() }
