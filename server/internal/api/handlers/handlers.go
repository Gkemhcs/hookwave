// Package handlers implements the HTTP handler layer: decoding requests,
// validating them, calling into the service layer, and mapping the result
// (or error) to an HTTP response.
package handlers

import (
	"encoding/json"
	"net/http"
)

// Handlers is the registry of all entity handlers, wired together in
// api.NewServer and consulted by the router.
type Handlers struct {
	Environment *EnvironmentHandler
	Application *ApplicationHandler
	Endpoint    *EndpointHandler
}

// NewHandlers builds a Handlers registry from already-constructed services.
func NewHandlers(environmentSvc EnvironmentService, applicationSvc ApplicationService, endpointSvc EndpointService) *Handlers {
	return &Handlers{
		Environment: &EnvironmentHandler{
			svc: environmentSvc,
		},
		Application: &ApplicationHandler{
			service: applicationSvc,
		},
		Endpoint: &EndpointHandler{
			service: endpointSvc,
		},
	}
}

// ErrorField is the error portion of an APIResponse.
type ErrorField struct {
	Msg         string `json:"msg"`
	Description string `json:"description,omitempty"`
	Code        int    `json:"code,omitempty"`
}

// APIResponse is the consistent envelope every handler responds with:
// either Data on success, or Err on failure, plus the request ID.
type APIResponse struct {
	Err       *ErrorField `json:"err,omitempty"`
	Data      any         `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// writeSuccessResponse writes data wrapped in an APIResponse with the given
// status code.
func writeSuccessResponse(statusCode int, w http.ResponseWriter, data any) error {
	requestID := w.Header().Get("X-Hookwave-Request-ID")
	response := APIResponse{
		Data:      data,
		RequestID: requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes an error, described by msg/desc, wrapped in an
// APIResponse with the given status code.
func writeErrorResponse(statusCode int, w http.ResponseWriter, msg, desc string) error {
	requestID := w.Header().Get("X-Hookwave-Request-ID")
	response := APIResponse{
		Err: &ErrorField{
			Msg:         msg,
			Description: desc,
			Code:        statusCode,
		},
		RequestID: requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)

}
