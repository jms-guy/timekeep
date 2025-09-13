//go:build !windows && !linux

package main

func (r *realServiceCommander) WriteToService() error {
	return nil
}
