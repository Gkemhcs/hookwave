package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/api/dto"
	"github.com/Gkemhcs/hookwave/server/internal/api/httpctx"
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging/tag"
	"github.com/Gkemhcs/hookwave/server/internal/service"
	"github.com/google/uuid"
)

// ApplicationService is the subset of *service.ApplicationService that
// ApplicationHandler depends on.
type ApplicationService interface {
	CreateApplication(ctx context.Context, args *service.CreateApplicationParams) (*domain.Application, error)
	GetApplicationByID(ctx context.Context, args *service.GetApplicationByIDParams) (*domain.Application, error)
	GetAllApplicationsByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*domain.Application, error)
	DeleteApplicationByID(ctx context.Context, args *service.DeleteApplicationByIDParams) error
	UpdateApplicationByID(ctx context.Context, args *service.UpdateApplicationByIDParams) (*domain.Application, error)
}

// ApplicationHandler handles the /environments/{environment_id}/applications
// HTTP routes.
type ApplicationHandler struct {
	service ApplicationService
}

// NewApplicationHandler builds an ApplicationHandler backed by service.
func NewApplicationHandler(service ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		service: service,
	}
}

// CreateApplication handles POST /environments/{environment_id}/applications.
func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)

	var req dto.CreateApplicationInput

	// Decode failure: malformed body - 400.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}
	id, err := uuid.Parse(r.PathValue("environment_id"))
	if err != nil {
		logger.Warn("environment id is missing or invalid uuid in url path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is missing or invalid")
		return
	}
	req.EnvironmentID = id
	// Validation failure: well-formed body, invalid values - 422, not 400.
	if err := req.Validate(); err != nil {
		logger.Warn("request body failed validation", tag.NewTagError(err))
		writeErrorResponse(http.StatusUnprocessableEntity, w, "VALIDATION_ERROR", err.Error())
		return
	}
	application, err := h.service.CreateApplication(ctx,
		&service.CreateApplicationParams{
			Name:          *req.Name,
			Slug:          *req.Slug,
			EnvironmentID: req.EnvironmentID,
		})
	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceAlreadyExist) {
				logger.Warn("resource already exists", tag.NewTagError(err))
				writeErrorResponse(http.StatusConflict, w, apiErr.Code, "application already exists")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("environment doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusBadRequest, w, apiErr.Code, "environment doesn't exist")
				return
			}
			logger.Error("unexpected application error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", application))
	writeSuccessResponse(http.StatusCreated, w, map[string]any{
		"application": application,
	})
}

// GetApplicationByID handles
// GET /environments/{environment_id}/applications/{application_id}.
func (h *ApplicationHandler) GetApplicationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environmentID, err := uuid.Parse(r.PathValue("environment_id"))
	if err != nil {
		logger.Warn("environment id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is not valid")
		return
	}

	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}
	application, err := h.service.GetApplicationByID(ctx, &service.GetApplicationByIDParams{
		EnvironmentID: environmentID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiError *apierror.APIError
		if ok := errors.As(err, &apiError); ok {
			if apiError == apierror.ErrResourceNotFound {
				logger.Warn("application not found in environment", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiError.Code, "application not found in environment")
				return
			}
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", application))
	writeSuccessResponse(http.StatusOK, w, application)

}

// UpdateApplicationByID handles
// PATCH /environments/{environment_id}/applications/{application_id}.
func (h *ApplicationHandler) UpdateApplicationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environmentID, err := uuid.Parse(r.PathValue("environment_id"))
	if err != nil {
		logger.Warn("environment id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is not valid")
		return
	}

	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}
	var req dto.UpdateApplicationInput
	// Decode failure: malformed body - 400.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}

	// Validation failure: well-formed body, invalid values - 422, not 400.
	if err := req.Validate(); err != nil {
		logger.Warn("request body failed validation", tag.NewTagError(err))
		writeErrorResponse(http.StatusUnprocessableEntity, w, "VALIDATION_ERROR", err.Error())
		return
	}

	application, err := h.service.UpdateApplicationByID(ctx, &service.UpdateApplicationByIDParams{
		ID:            applicationID,
		Name:          *req.Name,
		Slug:          *req.Slug,
		EnvironmentID: environmentID,
	})

	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceNotFound) {
				logger.Warn("application doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiErr.Code, "application doesn't exist")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("environment doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusBadRequest, w, apiErr.Code, "environment doesn't exist")
				return
			}
			logger.Error("unexpected application error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", application))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"application": application,
	})
}

// GetAllApplicationsByEnvironmentID handles
// GET /environments/{environment_id}/applications.
func (h *ApplicationHandler) GetAllApplicationsByEnvironmentID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environmentID, err := uuid.Parse(r.PathValue("environment_id"))
	if err != nil {
		logger.Warn("environment id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is not valid")
		return
	}
	applications, err := h.service.GetAllApplicationsByEnvironmentID(ctx, environmentID)
	if err != nil {
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", applications))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"applications": applications,
	})
}

// DeleteApplicationByID handles
// DELETE /environments/{environment_id}/applications/{application_id}.
func (h *ApplicationHandler) DeleteApplicationByID(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	environmentID, err := uuid.Parse(r.PathValue("environment_id"))
	if err != nil {
		logger.Warn("environment id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "environment id is not valid")
		return
	}

	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}
	err = h.service.DeleteApplicationByID(ctx, &service.DeleteApplicationByIDParams{
		ID:            environmentID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceNotFound) {
				logger.Warn("application not found", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiErr.Code, "application doesn't exist")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("application still has dependent resources, cannot delete", tag.NewTagError(err))
				writeErrorResponse(http.StatusConflict, w, apiErr.Code, "delete this application's endpoints first, then delete the application")
				return
			}
			logger.Error("unexpected application error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request")
	writeSuccessResponse(http.StatusOK, w, map[string]string{
		"message":        "application deleted successfully",
		"application_id": applicationID.String(),
		"environment_id": environmentID.String(),
	})

}
