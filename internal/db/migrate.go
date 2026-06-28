// This file runs database "migrations" on startup.

package db

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // registers "postgres://"
	_ "github.com/golang-migrate/migrate/v4/source/file"       // registers "file://"
)

// RunMigrations applies any pending migrations from the migrations/ folder
func RunMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
