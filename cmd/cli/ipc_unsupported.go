//go:build !windows && !linux

package main

func WriteToService() error {
	return nil
}
