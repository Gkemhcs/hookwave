package db

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies every pending up-migration found under
// migrationsPath to the database at dbURL. It is idempotent — running it
// again with nothing new to apply is not an error.
func RunMigrations(dbURL, migrationsPath string) error {
	// dbURL already carries its own scheme (postgres://...) for pgxpool's
	// benefit; golang-migrate's pgx5 driver instead selects itself by the
	// pgx5:// scheme, so swap the scheme rather than blindly prepending one
	// (which would double up and produce an unparseable URL).
	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse database url: %w", err)
	}
	u.Scheme = "pgx5"

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		u.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to initialise migrations: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
