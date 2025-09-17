//go:build !windows && !linux

package logs

func getLogPath() (string, error) {
	return "", nil
}
