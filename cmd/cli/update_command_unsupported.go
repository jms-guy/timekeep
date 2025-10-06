//go:build !windows && !linux

package main

func (s *CLIService) UpdateTimekeep() error {
	return nil
}
