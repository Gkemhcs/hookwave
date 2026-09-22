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


type MessageService interface {
	CreateMessage(ctx context.Context, args *service.CreateMessageParams) (*domain.Message, error)
}
type MessageHandler struct {
	service MessageService
}

func NewMessageHandler(service MessageService)*MessageHandler{
	return &MessageHandler{
		service: service,
	}
}

func(h *MessageHandler) Send(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	logger, _ := httpctx.LoggerFromContext(ctx)

	var req dto.CreateMessageInput
	// Decode failure: the request itself is malformed (bad JSON) - 400.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", err.Error())
		return
	}

	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		logger.Warn("application id is not valid in path", tag.NewTagError(err))
		writeErrorResponse(http.StatusBadRequest, w, "BAD_REQUEST", "application id is not valid")
		return
	}
	req.ApplicationID=applicationID
	// Validation failure: the request parsed fine, but the values in the
	// body don't satisfy the rules - 422 Unprocessable Entity, not 400.
	if err := req.Validate(); err != nil {
		logger.Warn("request body failed validation", tag.NewTagError(err))
		writeErrorResponse(http.StatusUnprocessableEntity, w, "VALIDATION_ERROR", err.Error())
		return
	}

	message,err:=h.service.CreateMessage(ctx,&service.CreateMessageParams{
		ApplicationID: req.ApplicationID,
		EventType: req.EventType,
		Payload: *req.Payload,
		Metadata: req.Metadata,
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
			logger.Error("unexpected message error", tag.NewTagError(err))
			writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
			return
		}
		logger.Error("failed to process the request", tag.NewTagError(err))
		writeErrorResponse(http.StatusInternalServerError, w, InternalServerError, InternalServerErrorDesc)
		return
	}
	logger.Info("processed the request", tag.NewTag("response", message))
	writeSuccessResponse(http.StatusAccepted, w, map[string]any{
		"message": message,
	})

}
