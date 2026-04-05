package database

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// InitializeDatabase initializes the database by running migrations
func InitializeDatabase(db *sql.DB) error {
	// Set the base filesystem for migrations
	goose.SetBaseFS(migrationsFS)

	// Run migrations
	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
