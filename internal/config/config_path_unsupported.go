//go:build !windows && !linux

package config

func getConfigLocation() (string, error) {
	return "", nil
}
