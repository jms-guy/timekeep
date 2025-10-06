//go:build !windows && !linux

package main

func (s *CLIService) StatusService() error {
	return nil
}
