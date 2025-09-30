//go:build linux

package sql

// Gets database directory path for Windows
func getDatabasePath() (string, error) {
	return "/var/lib/timekeep/timekeep.db", nil
}
