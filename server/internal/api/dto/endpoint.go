package dto

import (
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/google/uuid"
)

// CreateEndpointInput is the body of
// POST /applications/{application_id}/endpoints.
type CreateEndpointInput struct {
	Name          *string        `json:"name"`
	ApplicationID uuid.UUID      `json:"application_id"`
	URL           *string        `json:"url"`
	Method        string         `json:"method"`
	Headers       map[string]any `json:"headers"`
	QueryParams   map[string]any `json:"query_params"`
	EventTypes    []string       `json:"event_types"`
	SigningSecret string         `json:"signing_secret"`
	TimeoutMs     int32          `json:"timeout_ms"`
}

// Validate checks that Name and URL were provided.
func (req *CreateEndpointInput) Validate() error {
	if req.Name == nil || *req.Name=="" {
		return apierror.NewValidationError("endpoint name should not be empty")
	}
	if req.URL == nil || *req.URL==""{
		return apierror.NewValidationError("endpoint url should not be empty")
	}
	if req.TimeoutMs<0{
		return apierror.NewValidationError("timeout ms should not be negative, it should be greater than or equal to 0")
	}
	return nil
}

// UpdateEndpointByIDInput is the body of
// PATCH /applications/{application_id}/endpoints/{endpoint_id}. Every field
// except ID/ApplicationID is optional — a nil field is left unchanged.
type UpdateEndpointByIDInput struct {
	Name          *string         `json:"name"`
	URL           *string         `json:"url"`
	Method        *string         `json:"method"`
	Headers       *map[string]any `json:"headers"`
	QueryParams   *map[string]any `json:"query_params"`
	EventTypes    *[]string       `json:"event_types"`
	SigningSecret *string         `json:"signing_secret"`
	TimeoutMs     *int32          `json:"timeout_ms"`
	ID            uuid.UUID       `json:"id"`
	ApplicationID uuid.UUID       `json:"application_id"`
}

// Validate checks that, if provided, Name isn't empty and TimeoutMs isn't
// negative.
func (req *UpdateEndpointByIDInput) Validate() error {
	if req.Name != nil && *req.Name == "" {
		return apierror.NewValidationError("endpoint name cannot be set emtpy")
	}
	if req.TimeoutMs != nil && *req.TimeoutMs < 0 {
		return apierror.NewValidationError("timeout ms should not be negative , it should be greater than equal to 0")
	}
	return nil
}
