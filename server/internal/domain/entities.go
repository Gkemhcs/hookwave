// Package domain holds Hookwave's plain entity types — the shapes returned
// up through the service layer to handlers, independent of how they're
// stored (repository types) or how they arrived over HTTP (dto types).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Environment is the top-level tenant boundary — every application lives
// under exactly one environment.
type Environment struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Application belongs to one Environment and is the unit event messages and
// endpoints are scoped to.
type Application struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Endpoint is a URL registered under an Application to receive webhook
// deliveries for the event types it subscribes to.
type Endpoint struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	ApplicationID uuid.UUID      `json:"application_id"`
	URL           string         `json:"url"`
	Method        string         `json:"method"`
	Headers       map[string]any `json:"headers"`
	QueryParams   map[string]any `json:"query_params"`
	EventTypes    []string       `json:"event_types"`
	SigningSecret string         `json:"signing_secret"`
	IsActive      bool           `json:"is_active"`
	TimeoutMs     int32          `json:"timeout_ms"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
