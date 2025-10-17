package sql

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jms-guy/timekeep/internal/database"
	"github.com/pressly/goose/v3"
)

//go:embed schema/*.sql
var embedMigrations embed.FS

// Open database connection with embedded migrations
func OpenLocalDatabase() (*database.Queries, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}

	// #nosec G301
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
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

// Opens functional in-memory testing database
func OpenTestDatabase() (*database.Queries, error) {
	db, err := sql.Open("sqlite", ":memory:")
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
