package db

import (
	"database/sql"
	"fmt"
	"simple-procmon/config"

	"github.com/pressly/goose/v3"
)

func Migrate(database *sql.DB) error {

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("could not set dialect: %w", err)
	}

	if err := goose.Up(database, config.GlobalConfig.MigrationsDir); err != nil {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	return nil
}
