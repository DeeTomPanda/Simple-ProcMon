package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"simple-procmon/internal/constants"

	_ "modernc.org/sqlite"
)

func Open() (*sql.DB, error) {

	dbPath := os.Getenv(constants.DBEnvKey)
	if dbPath == "" {
		if err := os.MkdirAll(constants.DefaultDBDir, 0o755); err != nil {
			log.Fatalf("failed to create db dir: %v", err)
		}
	}

	dbPath = filepath.Join(dbPath, constants.DefaultDBFile)

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("could not create db directory: %w", err)
	}

	database, err := sql.Open(constants.DBDriver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("could not secure db file: %w", err)
	}

	if _, err := database.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("could not set WAL mode: %w", err)
	}

	return database, nil
}
