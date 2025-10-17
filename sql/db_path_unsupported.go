//go:build !windows && !linux

package sql

func getDatabasePath() (string, error) {
	return "", nil
}
