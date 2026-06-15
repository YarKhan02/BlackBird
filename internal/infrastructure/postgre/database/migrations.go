package database

import (
	"fmt"

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

	m, err := migrate.New(
		sourceURL,
		databaseURL,
	)

	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
