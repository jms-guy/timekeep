package internal

import (
	"fmt"

	mySQL "github.com/jms-guy/timekeep/sql"
)

func LoadConfig() error {
	localDb := "timekeep.db"

	_, err := mySQL.OpenLocalDatabase(localDb)
	if err != nil {
		return fmt.Errorf("error opening local database connection: %w", err)
	}

	return nil
}
