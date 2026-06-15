package database

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigration(sourceURL, databaseURL string) error {
	if sourceURL == "" {
		sourceURL = "file://./migrations"
	}
	if databaseURL == "" {
		return fmt.Errorf("database URL is empty")
	}
	databaseURL = normalizeMigrateDatabaseURL(databaseURL)

	m, err := migrate.New(
		sourceURL,
		databaseURL,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func normalizeMigrateDatabaseURL(raw string) string {
	// golang-migrate with pgx expects pgx:// style DSNs.
	if strings.HasPrefix(raw, "postgresql://") {
		return "pgx://" + strings.TrimPrefix(raw, "postgresql://")
	}
	if strings.HasPrefix(raw, "postgres://") {
		return "pgx://" + strings.TrimPrefix(raw, "postgres://")
	}
	return raw
}
