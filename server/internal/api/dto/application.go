package dto

import (
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/google/uuid"
)

// CreateApplicationInput is the body of
// POST /environments/{environment_id}/applications.
type CreateApplicationInput struct {
	Name          *string   `json:"name"`
	Slug          *string   `json:"slug"`
	EnvironmentID uuid.UUID `json:"environment_id"`
}

// GetApplicationByIDInput identifies one application by its environment and
// application IDs.
type GetApplicationByIDInput struct {
	EnvironmentID uuid.UUID `json:"environment_id"`
	ApplicationID uuid.UUID `json:"application_id"`
}

// Validate checks that Name and Slug were provided and non-empty.
func (req *CreateApplicationInput) Validate() error {
	if req.Name == nil || *req.Name == "" {
		return apierror.NewValidationError("name should not be empty")
	}
	if req.Slug == nil || *req.Slug == "" {
		return apierror.NewValidationError("slug should not be empty")
	}
	return nil
}

// UpdateApplicationInput is the body of
// PATCH /environments/{environment_id}/applications/{application_id}. Name
// and Slug are optional — a nil field is left unchanged.
type UpdateApplicationInput struct {
	ID            uuid.UUID `json:"id"`
	Name          *string   `json:"name"`
	Slug          *string   `json:"slug"`
	EnvironmentID uuid.UUID `json:"environment_id"`
}

// Validate currently performs no checks.
func (req *UpdateApplicationInput) Validate() error {
	return nil
}
