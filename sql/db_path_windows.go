// go: build windows
package sql

import (
	"log"
	"path/filepath"
)

// Gets database directory path for Windows
func getDatabasePath(logger *log.Logger) (string, error) {
	dataDir := `C:\ProgramData\TimeKeep`
	return filepath.Join(dataDir, "timekeep.db"), nil
}
