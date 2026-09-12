package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// defaultHTTPMethod is used when an endpoint is created without an
	// explicit HTTP method.
	defaultHTTPMethod = "POST"
	// defaultEndpointTimeoutMs is used when an endpoint is created without
	// an explicit timeout.
	defaultEndpointTimeoutMs = 5000
	active                   = true
	inactive                 = false
)

// EndpointRepository is the subset of the generated repository.Querier
// that EndpointService depends on.
type EndpointRepository interface {
	CreateEndpoint(ctx context.Context, arg repository.CreateEndpointParams) (repository.Endpoint, error)
	DeleteEndpointByID(ctx context.Context, arg repository.DeleteEndpointByIDParams) error
	GetAllEndpointsByApplicationID(ctx context.Context, applicationID uuid.UUID) ([]repository.Endpoint, error)
	GetEndpointByID(ctx context.Context, arg repository.GetEndpointByIDParams) (repository.Endpoint, error)
	UpdateEndpointByID(ctx context.Context, arg repository.UpdateEndpointByIDParams) (repository.Endpoint, error)
	UpdateEndpointStatus(ctx context.Context, arg repository.UpdateEndpointStatusParams) (repository.Endpoint, error)
}

// CreateEndpointParams is the input to CreateEndpoint. Method and TimeoutMs
// fall back to their defaults when left zero-valued.
type CreateEndpointParams struct {
	Name          string
	ApplicationID uuid.UUID
	URL           string
	Method        string
	Headers       map[string]any
	QueryParams   map[string]any
	EventTypes    []string
	SigningSecret string
	TimeoutMs     int32
}

// EndpointByIDParams identifies one endpoint by its application and
// endpoint IDs. Used by every operation that doesn't also need a body.
type EndpointByIDParams struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
}

// UpdateEndpointByIDParams is the input to UpdateEndpointByID. Name/URL use
// an empty string to mean "leave unchanged"; every other optional field
// uses a nil pointer for the same purpose.
type UpdateEndpointByIDParams struct {
	Name          string
	URL           string
	Method        *string
	Headers       *map[string]any
	QueryParams   *map[string]any
	EventTypes    *[]string
	SigningSecret *string
	TimeoutMs     *int32
	ID            uuid.UUID
	ApplicationID uuid.UUID
}

// NewEndpointService builds an EndpointService backed by repository.
func NewEndpointService(repository EndpointRepository) *EndpointService {
	return &EndpointService{
		repository: repository,
	}
}

// EndpointService implements endpoint business logic.
type EndpointService struct {
	repository EndpointRepository
}

// CreateEndpoint creates a new endpoint under an application.
func (s *EndpointService) CreateEndpoint(ctx context.Context, args *CreateEndpointParams) (*domain.Endpoint, error) {
	currentTime := time.Now().UTC()
	endpointParams := repository.CreateEndpointParams{
		ID:            uuid.New(),
		Name:          args.Name,
		ApplicationID: args.ApplicationID,
		Url:           args.URL,
		Method:        s.getEffectiveMethod(args.Method),
		EventTypes:    args.EventTypes,
		IsActive:      active,
		TimeoutMs:     s.getEffectiveTimeout(args.TimeoutMs),
		CreatedAt:     currentTime,
		UpdatedAt:     currentTime,
	}
	if args.SigningSecret == "" {
		endpointParams.SigningSecret = pgtype.Text{
			Valid: false,
		}
	} else {
		endpointParams.SigningSecret = pgtype.Text{
			String: args.SigningSecret,
			Valid:  true,
		}
	}
	queryParams, err := json.Marshal(args.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid query params in request body")
	}
	endpointParams.QueryParams = (*json.RawMessage)(&queryParams)

	headers, err := json.Marshal(args.Headers)
	if err != nil {
		return nil, fmt.Errorf("invalid headers in request body")
	}
	endpointParams.Headers = (*json.RawMessage)(&headers)
	endpoint, err := s.repository.CreateEndpoint(ctx, endpointParams)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildEndpoint(&endpoint)
}

// GetEndpointByID returns the endpoint identified by args, scoped to its
// application.
func (s *EndpointService) GetEndpointByID(ctx context.Context, args *EndpointByIDParams) (*domain.Endpoint, error) {
	endpoint, err := s.repository.GetEndpointByID(ctx, repository.GetEndpointByIDParams{
		ID:            args.ID,
		ApplicationID: args.ApplicationID,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildEndpoint(&endpoint)
}

// GetAllEndpointsByApplicationID returns every endpoint under
// applicationID.
func (s *EndpointService) GetAllEndpointsByApplicationID(ctx context.Context, applicationID uuid.UUID) ([]*domain.Endpoint, error) {

	endpointsFromRepo, err := s.repository.GetAllEndpointsByApplicationID(ctx, applicationID)
	if err != nil {
		return []*domain.Endpoint{}, repository.MapError(err)
	}
	var endpoints []*domain.Endpoint
	for _, endpointFromRepo := range endpointsFromRepo {
		endpoint, err := s.buildEndpoint(&endpointFromRepo)
		if err == nil {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints, nil
}

// DeleteEndpointByID deletes the endpoint identified by args, scoped to its
// application.
func (s *EndpointService) DeleteEndpointByID(ctx context.Context, args *EndpointByIDParams) error {
	err := s.repository.DeleteEndpointByID(ctx, repository.DeleteEndpointByIDParams{
		ID:            args.ID,
		ApplicationID: args.ApplicationID,
	})
	if err != nil {
		return repository.MapError(err)
	}
	return nil
}

// UpdateEndpointByID applies a partial update to the endpoint identified by
// args, scoped to its application. A zero-valued/nil field in args leaves
// the existing value unchanged.
func (s *EndpointService) UpdateEndpointByID(ctx context.Context, args *UpdateEndpointByIDParams) (*domain.Endpoint, error) {
	endpoint, err := s.repository.GetEndpointByID(ctx, repository.GetEndpointByIDParams{
		ApplicationID: args.ApplicationID,
		ID:            args.ID,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	endpointParams := repository.UpdateEndpointByIDParams{
		ID:            endpoint.ID,
		Name:          endpoint.Name,
		ApplicationID: endpoint.ApplicationID,
		Url:           endpoint.Url,
		Method:        endpoint.Method,
		EventTypes:    endpoint.EventTypes,
		IsActive:      endpoint.IsActive,
		TimeoutMs:     endpoint.TimeoutMs,
		SigningSecret: endpoint.SigningSecret,
		QueryParams:   endpoint.QueryParams,
		Headers:       endpoint.Headers,
	}
	if args.Name != "" {
		endpointParams.Name = args.Name
	}
	if args.URL != "" {
		endpointParams.Url = args.URL
	}
	if args.Method != nil {
		endpointParams.Method = *args.Method
	}
	if args.EventTypes != nil {
		endpointParams.EventTypes = *args.EventTypes
	}
	if args.TimeoutMs != nil {
		endpointParams.TimeoutMs = *args.TimeoutMs
	}
	if args.SigningSecret != nil {
		if *args.SigningSecret == "" {
			endpointParams.SigningSecret = pgtype.Text{
				Valid: false,
			}
		} else {
			endpointParams.SigningSecret = pgtype.Text{
				String: *args.SigningSecret,
				Valid:  true,
			}
		}
	}
	if args.QueryParams != nil {
		queryParams, err := json.Marshal(args.QueryParams)
		if err != nil {
			return nil, fmt.Errorf("invalid query params in request body")
		}
		endpointParams.QueryParams = (*json.RawMessage)(&queryParams)
	}

	if args.Headers != nil {
		headers, err := json.Marshal(args.Headers)
		if err != nil {
			return nil, fmt.Errorf("invalid headers in request body")
		}
		endpointParams.Headers = (*json.RawMessage)(&headers)
	}

	endpoint, err = s.repository.UpdateEndpointByID(ctx, endpointParams)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildEndpoint(&endpoint)
}

// DisableEndpoint sets the endpoint identified by args inactive, scoped to
// its application.
func (s *EndpointService) DisableEndpoint(ctx context.Context, args *EndpointByIDParams) (*domain.Endpoint, error) {
	endpoint, err := s.repository.UpdateEndpointStatus(ctx, repository.UpdateEndpointStatusParams{
		ID:            args.ID,
		ApplicationID: args.ApplicationID,
		IsActive:      inactive,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildEndpoint(&endpoint)
}

// EnableEndpoint sets the endpoint identified by args active, scoped to its
// application.
func (s *EndpointService) EnableEndpoint(ctx context.Context, args *EndpointByIDParams) (*domain.Endpoint, error) {
	endpoint, err := s.repository.UpdateEndpointStatus(ctx, repository.UpdateEndpointStatusParams{
		ID:            args.ID,
		ApplicationID: args.ApplicationID,
		IsActive:      active,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildEndpoint(&endpoint)
}

// buildEndpoint converts a repository row into the domain type, parsing the
// stored JSONB headers/query params.
func (s *EndpointService) buildEndpoint(endpoint *repository.Endpoint) (*domain.Endpoint, error) {

	var headers map[string]any
	if endpoint.Headers != nil {
		if err := json.Unmarshal(*endpoint.Headers, &headers); err != nil {
			return nil, fmt.Errorf("parse headers: %w", err)
		}
	}

	var queryParams map[string]any
	if endpoint.QueryParams != nil {
		if err := json.Unmarshal(*endpoint.QueryParams, &queryParams); err != nil {
			return nil, fmt.Errorf("parse query params: %w", err)
		}
	}
	return &domain.Endpoint{
		ID:            endpoint.ID,
		Name:          endpoint.Name,
		URL:           endpoint.Url,
		ApplicationID: endpoint.ApplicationID,
		Method:        endpoint.Method,
		Headers:       headers,
		QueryParams:   queryParams,
		EventTypes:    endpoint.EventTypes,
		SigningSecret: endpoint.SigningSecret.String,
		IsActive:      endpoint.IsActive,
		TimeoutMs:     endpoint.TimeoutMs,
		CreatedAt:     endpoint.CreatedAt,
		UpdatedAt:     endpoint.UpdatedAt,
	}, nil
}

// getEffectiveMethod returns method, or defaultHTTPMethod if method is
// empty.
func (s *EndpointService) getEffectiveMethod(method string) string {
	if method == "" {
		return defaultHTTPMethod
	}
	return method
}

// getEffectiveTimeout returns timeout, or defaultEndpointTimeoutMs if
// timeout is zero.
func (s *EndpointService) getEffectiveTimeout(timeout int32) int32 {
	if timeout == 0 {
		return defaultEndpointTimeoutMs
	}
	return timeout
}
