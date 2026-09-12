// Command migrate is a small CLI entrypoint that applies all pending
// golang-migrate migrations from the migrations directory against the
// database configured via HOOKWAVE_DATABASE_CONNECTION_STRING.
package main

import (
	"fmt"
	"os"

	"github.com/Gkemhcs/hookwave/server/internal/config"
	"github.com/Gkemhcs/hookwave/server/internal/db"
)

func main() {
	serverConfig := config.New()
	if err := serverConfig.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "config loading failed: %v\n", err)
		os.Exit(1)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	if err := db.RunMigrations(serverConfig.DBConn, migrationsPath); err != nil {
		fmt.Fprintf(os.Stderr, "migrations failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("migrations applied successfully")
}
