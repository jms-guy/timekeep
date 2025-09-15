//go:build !linux

package main

import "fmt"

type windowsDaemonNoop struct{}

func NewDaemonManager() (DaemonManager, error) {
	return &windowsDaemonNoop{}, nil
}

func (w *windowsDaemonNoop) Install() (string, error) {
	return "N/A", fmt.Errorf("daemon management not supported from this binary")
}
func (w *windowsDaemonNoop) Remove() (string, error) {
	return "N/A", fmt.Errorf("daemon management not supported from this binary")
}
func (w *windowsDaemonNoop) Start() (string, error) {
	return "N/A", fmt.Errorf("daemon management not supported from this binary")
}
func (w *windowsDaemonNoop) Stop() (string, error) {
	return "N/A", fmt.Errorf("daemon management not supported from this binary")
}
func (w *windowsDaemonNoop) Status() (string, error) {
	return "N/A", fmt.Errorf("daemon management not supported from this binary")
}
