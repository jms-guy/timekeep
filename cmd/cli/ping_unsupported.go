//go:build !windows && !linux

package main

func (s *CLIService) PingService() error {
	return nil
}
