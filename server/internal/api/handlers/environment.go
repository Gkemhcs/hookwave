package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/api/dto"
	"github.com/Gkemhcs/hookwave/server/internal/api/httpctx"
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging/tag"
	"github.com/Gkemhcs/hookwave/server/internal/service"
	"github.com/google/uuid"
)

// EnvironmentService is the subset of *service.EnvironmentService that
// EnvironmentHandler depends on.
type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, arg *service.CreateEnvironmentParams) (*domain.Environment, error)
	DeleteEnvironment(ctx context.Context, arg *service.DeleteEnvironmentParams) error
	ListEnvironments(ctx context.Context) ([]domain.Environment, error)
	GetEnvironmentByID(ctx context.Context, environmentID uuid.UUID) (*domain.Environment, error)
}

// EnvironmentHandler handles the /environments HTTP routes.
type EnvironmentHandler struct {
	svc EnvironmentService
}

// NewEnvironmentHandler builds an EnvironmentHandler backed by service.
func NewEnvironmentHandler(service EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{
		svc: service,
	}
}

// CreateEnvironment handles POST /environments.
func (h *EnvironmentHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)

	var req dto.CreateEnvironmentRequest

	// Decode failure: the request itself is malformed (bad JSON) - 400.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}

	// Validation failure: the request parsed fine, but the values in the
	// body don't satisfy the rules - 422 Unprocessable Entity, not 400.
	if err := req.Validate(); err != nil {
		logger.Warn("request body failed validation", tag.NewTagError(err))
		writeErrorResponse(http.StatusUnprocessableEntity, w, "VALIDATION_ERROR", err.Error())
		return
	}
	environment, err := h.svc.CreateEnvironment(ctx, &service.CreateEnvironmentParams{
		Name: *req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {

			if apiErr == apierror.ErrResourceAlreadyExist {
				logger.Warn("resource already exists", tag.NewTagError(err))
				writeErrorResponse(http.StatusConflict, w, apiErr.Code, "environment already exists")
				return
			}
			logger.Error("unexpected application error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, apiErr.Code, apiErr.Message)
			return

		}
		logger.Error("failed to process request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", environment))
	writeSuccessResponse(http.StatusCreated, w, environment)

}

// DeleteEnvironment handles DELETE /environments/{id}.
func (h *EnvironmentHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)

	var req dto.DeleteEnvironmentRequest
	environmentId := r.PathValue("id")
	req.ID = &environmentId

	// This validates a path parameter, not a request body - stays 400,
	// not 422 (422 is specifically for a malformed/invalid body).
	if err := req.Validate(); err != nil {
		logger.Warn("invalid path parameter", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}
	err := h.svc.DeleteEnvironment(ctx, &service.DeleteEnvironmentParams{
		Id: uuid.MustParse(*req.ID),
	})
	if err != nil {
		if errors.Is(err, apierror.ErrResourceNotFound) {
			logger.Warn("environment not found", tag.NewTagError(err))
			writeErrorResponse(http.StatusNotFound, w, "RESOURCE_NOT_FOUND", EnvironmentNotFound)
			return
		}
		logger.Error("failed to process request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	response := map[string]string{
		"message": fmt.Sprintf("environment %s deleted successfully", *req.ID),
	}
	logger.Info("processed the request", tag.NewTag("response", response))
	writeSuccessResponse(http.StatusOK, w, response)

}

// ListEnvironments handles GET /environments.
func (h *EnvironmentHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environments, err := h.svc.ListEnvironments(ctx)
	if err != nil {
		logger.Error("failed to process request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", environments))

	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"environments": environments,
	})

}

// GetEnvironmentById handles GET /environments/{id}.
func (h *EnvironmentHandler) GetEnvironmentById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environmentId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Warn("environment id is missing or invalid uuid in url path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is missing or invalid")
		return
	}
	environment, err := h.svc.GetEnvironmentByID(ctx, environmentId)
	if err != nil {
		if errors.Is(err, apierror.ErrResourceNotFound) {
			logger.Warn("environment not found", tag.NewTagError(err))
			writeErrorResponse(http.StatusNotFound, w, "RESOURCE_NOT_FOUND", "environment not found")
			return
		}
		logger.Error("failed to process request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}

	logger.Info("processed the request", tag.NewTag("response", environment))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"environment": environment,
	})

}
