// go: build windows
package sql

import "path/filepath"

// Gets database directory path for Windows
func getDatabasePath() (string, error) {
	dataDir := `C:\ProgramData\TimeKeep`
	return filepath.Join(dataDir, "timekeep.db"), nil
}
