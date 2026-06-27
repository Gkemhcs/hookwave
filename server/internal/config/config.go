package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port   int    `env:"PORT" envDefault:"8080"`
	DBConn string `env:"HOOKWAVE_DATABASE_CONNECTION_STRING" envDefault:"postgres://postgres@localhost:5432/postgres"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

func (sc *ServerConfig) Load() error {
	godotenv.Load()

	err := env.Parse(sc)
	if err != nil {
		return err
	}
	return nil
}
