// Package dto holds the request-body shapes handlers decode into, plus
// their Validate methods. Each type is layer-specific to HTTP — the
// service layer never depends on these.
package dto

import (
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/google/uuid"
)

// CreateEnvironmentRequest is the body of POST /environments.
type CreateEnvironmentRequest struct {
	Name *string `json:"name"`
	Slug string  `json:"slug"`
}

// Validate checks that Name was provided.
func (req *CreateEnvironmentRequest) Validate() error {
	if req.Name == nil {
		return apierror.NewValidationError("name is empty")

	}
	return nil
}

// DeleteEnvironmentRequest carries the id path parameter for
// DELETE /environments/{id}.
type DeleteEnvironmentRequest struct {
	ID *string `json:"id"`
}

// Validate checks that Id was provided and is a valid UUID.
func (req *DeleteEnvironmentRequest) Validate() error {
	if req.ID == nil {
		return apierror.NewValidationError("id is empty")
	}
	if _, err := uuid.Parse(*req.ID); err != nil {
		return apierror.NewValidationError("invalid id , must be uuid")
	}
	return nil
}
