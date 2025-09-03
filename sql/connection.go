package sql

import (
	"database/sql"
	"embed"
	"io"
	"log"

	"github.com/jms-guy/timekeep/internal/database"
	"github.com/pressly/goose/v3"
)

//go:embed schema/*.sql
var embedMigrations embed.FS

func OpenLocalDatabase() (*database.Queries, error) {
	localDb := "timekeep.db"

	db, err := sql.Open("sqlite", localDb)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log.New(io.Discard, "", 0))

	if err = goose.SetDialect("sqlite"); err != nil {
		return nil, err
	}

	if err = goose.Up(db, "schema"); err != nil {
		return nil, err
	}

	queries := database.New(db)

	return queries, nil
}
