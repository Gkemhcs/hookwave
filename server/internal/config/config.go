// Package config loads Hookwave's server configuration from environment
// variables (optionally via a local .env file).
package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// ServerConfig holds the environment-derived settings the server needs to
// start: the port to listen on and the Postgres connection string.
type ServerConfig struct {
	Port   int    `env:"PORT" envDefault:"8080"`
	DBConn string `env:"HOOKWAVE_DATABASE_CONNECTION_STRING" envDefault:"postgres://postgres@localhost:5432/postgres"`
}

// New returns a zero-value ServerConfig. Call Load to populate it.
func New() *ServerConfig {
	return &ServerConfig{}
}

// Load populates sc from environment variables, first loading a local .env
// file if one is present (missing .env is not an error).
func (sc *ServerConfig) Load() error {
	godotenv.Load()

	err := env.Parse(sc)
	if err != nil {
		return err
	}
	return nil
}
