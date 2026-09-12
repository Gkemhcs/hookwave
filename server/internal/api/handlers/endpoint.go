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

// EndpointService is the subset of *service.EndpointService that
// EndpointHandler depends on.
type EndpointService interface {
	CreateEndpoint(ctx context.Context, args *service.CreateEndpointParams) (*domain.Endpoint, error)
	GetEndpointByID(ctx context.Context, args *service.EndpointByIDParams) (*domain.Endpoint, error)
	GetAllEndpointsByApplicationID(ctx context.Context, applicationID uuid.UUID) ([]*domain.Endpoint, error)
	DeleteEndpointByID(ctx context.Context, args *service.EndpointByIDParams) error
	UpdateEndpointByID(ctx context.Context, args *service.UpdateEndpointByIDParams) (*domain.Endpoint, error)
	DisableEndpoint(ctx context.Context, args *service.EndpointByIDParams) (*domain.Endpoint, error)
	EnableEndpoint(ctx context.Context, args *service.EndpointByIDParams) (*domain.Endpoint, error)
}

// NewEndpointHandler builds an EndpointHandler backed by service.
func NewEndpointHandler(service EndpointService) *EndpointHandler {
	return &EndpointHandler{
		service: service,
	}
}

// EndpointHandler handles the
// /applications/{application_id}/endpoints HTTP routes.
type EndpointHandler struct {
	service EndpointService
}

// CreateEndpoint handles POST /applications/{application_id}/endpoints.
func (h *EndpointHandler) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)

	var req *dto.CreateEndpointInput

	// Decode failure: malformed body - 400.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}
	id, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is missing or invalid uuid in url path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is missing or invalid")
		return
	}
	req.ApplicationID = id
	// Validation failure: well-formed body, invalid values - 422, not 400.
	if err := req.Validate(); err != nil {
		logger.Warn("request body failed validation", tag.NewTagError(err))
		writeErrorResponse(http.StatusUnprocessableEntity, w, "VALIDATION_ERROR", err.Error())
		return
	}
	endpoint, err := h.service.CreateEndpoint(ctx,
		&service.CreateEndpointParams{
			Name:          *req.Name,
			ApplicationID: req.ApplicationID,
			URL:           *req.URL,
			Method:        req.Method,
			Headers:       req.Headers,
			QueryParams:   req.QueryParams,
			EventTypes:    req.EventTypes,
			SigningSecret: req.SigningSecret,
			TimeoutMs:     req.TimeoutMs,
		})
	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceAlreadyExist) {
				logger.Warn("resource already exists", tag.NewTagError(err))
				writeErrorResponse(http.StatusConflict, w, apiErr.Code, "endpoint already exists")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("application doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusBadRequest, w, apiErr.Code, "application doesn't exist")
				return
			}
			logger.Error("unexpected endpoint error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoint))
	writeSuccessResponse(http.StatusCreated, w, map[string]any{
		"endpoint": endpoint,
	})
}

// GetEndpointByID handles
// GET /applications/{application_id}/endpoints/{endpoint_id}.
func (h *EndpointHandler) GetEndpointByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}

	endpointID, err := uuid.Parse(r.PathValue("endpoint_id"))
	if err != nil {
		logger.Warn("endpoint id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "endpoint id is not valid")
		return
	}
	endpoint, err := h.service.GetEndpointByID(ctx, &service.EndpointByIDParams{
		ID:            endpointID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiError *apierror.APIError
		if ok := errors.As(err, &apiError); ok {
			if apiError == apierror.ErrResourceNotFound {
				logger.Warn("endpoint not found in environment", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiError.Code, "endpoint not found in applications")
				return
			}
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoint))
	writeSuccessResponse(http.StatusOK, w, endpoint)
}

// UpdateEndpointByID handles
// PATCH /applications/{application_id}/endpoints/{endpoint_id}.
func (h *EndpointHandler) UpdateEndpointByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}

	endpointID, err := uuid.Parse(r.PathValue("endpoint_id"))
	if err != nil {
		logger.Warn("endpoint id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "endpoint id is not valid")
		return
	}
	var req *dto.UpdateEndpointByIDInput
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

	// req.Name/req.URL are *string (nil means "not provided" in a partial
	// update); service.UpdateEndpointByIDParams wants plain strings, so a
	// missing field defaults to "" here, matching how the service already
	// treats an empty Name/URL as "leave this field unchanged".
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	url := ""
	if req.URL != nil {
		url = *req.URL
	}
	endpoint, err := h.service.UpdateEndpointByID(ctx, &service.UpdateEndpointByIDParams{
		Name:          name,
		URL:           url,
		Method:        req.Method,
		Headers:       req.Headers,
		QueryParams:   req.QueryParams,
		EventTypes:    req.EventTypes,
		SigningSecret: req.SigningSecret,
		TimeoutMs:     req.TimeoutMs,
		ID:            endpointID,
		ApplicationID: applicationID,
	})

	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceNotFound) {
				logger.Warn("endpoint doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiErr.Code, "endpoint doesn't exist")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("application doesn't exist", tag.NewTagError(err))
				writeErrorResponse(http.StatusBadRequest, w, apiErr.Code, "application doesn't exist")
				return
			}
			logger.Error("unexpected endpoint error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoint))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"endpoint": endpoint,
	})
}

// GetAllEndpointsByApplicationID handles
// GET /applications/{application_id}/endpoints.
func (h *EndpointHandler) GetAllEndpointsByApplicationID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}
	endpoints, err := h.service.GetAllEndpointsByApplicationID(ctx, applicationID)
	if err != nil {
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoints))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"endpoints": endpoints,
	})
}

// DeleteEndpointByID handles
// DELETE /applications/{application_id}/endpoints/{endpoint_id}.
func (h *EndpointHandler) DeleteEndpointByID(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}

	endpointID, err := uuid.Parse(r.PathValue("endpoint_id"))
	if err != nil {
		logger.Warn("endpoint id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "endpoint id is not valid")
		return
	}
	err = h.service.DeleteEndpointByID(ctx, &service.EndpointByIDParams{
		ID:            endpointID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiErr *apierror.APIError
		if ok := errors.As(err, &apiErr); ok {
			if errors.Is(apiErr, apierror.ErrResourceNotFound) {
				logger.Warn("endpoint not found", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiErr.Code, "endpoint doesn't exist")
				return
			}
			if errors.Is(apiErr, apierror.ErrParentNotExist) {
				logger.Warn("endpoint still has dependent resources, cannot delete", tag.NewTagError(err))
				writeErrorResponse(http.StatusConflict, w, apiErr.Code, "delete this endpoints's child resources  first, then delete the endpoint")
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
		"message":        "endpoint deleted successfully",
		"application_id": applicationID.String(),
		"endpoint_id":    endpointID.String(),
	})

}

// DisableEndpointByID handles
// PATCH /applications/{application_id}/endpoints/{endpoint_id}/disable.
func (h *EndpointHandler) DisableEndpointByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}

	endpointID, err := uuid.Parse(r.PathValue("endpoint_id"))
	if err != nil {
		logger.Warn("endpoint id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "endpoint id is not valid")
		return
	}
	endpoint, err := h.service.DisableEndpoint(ctx, &service.EndpointByIDParams{
		ID:            endpointID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiError *apierror.APIError
		if ok := errors.As(err, &apiError); ok {
			if apiError == apierror.ErrResourceNotFound {
				logger.Warn("endpoint not found in environment", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiError.Code, "endpoint not found in applications")
				return
			}
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoint))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"message":  "endpoint disabled successfully",
		"endpoint": endpoint,
	})
}

// EnableEndpointByID handles
// PATCH /applications/{application_id}/endpoints/{endpoint_id}/enable.
func (h *EndpointHandler) EnableEndpointByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}

	endpointID, err := uuid.Parse(r.PathValue("endpoint_id"))
	if err != nil {
		logger.Warn("endpoint id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "endpoint id is not valid")
		return
	}
	endpoint, err := h.service.EnableEndpoint(ctx, &service.EndpointByIDParams{
		ID:            endpointID,
		ApplicationID: applicationID,
	})
	if err != nil {
		var apiError *apierror.APIError
		if ok := errors.As(err, &apiError); ok {
			if apiError == apierror.ErrResourceNotFound {
				logger.Warn("endpoint not found in environment", tag.NewTagError(err))
				writeErrorResponse(http.StatusNotFound, w, apiError.Code, "endpoint not found in applications")
				return
			}
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", endpoint))
	writeSuccessResponse(http.StatusOK, w, map[string]any{
		"message":  "endpoint enabled successfully",
		"endpoint": endpoint,
	})
}
