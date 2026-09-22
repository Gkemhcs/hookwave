// Package service is the business-logic layer: it takes its own
// hand-owned params types (never dto or repository types on its public
// API), talks to a repository through a small interface defined here, and
// returns domain types.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/google/uuid"
)

// EnvironmentRepository is the subset of the generated repository.Querier
// that EnvironmentService depends on.
type EnvironmentRepository interface {
	CreateEnvironment(ctx context.Context, arg repository.CreateEnvironmentParams) (repository.Environment, error)
	DeleteEnvironment(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ListEnvironments(ctx context.Context) ([]repository.Environment, error)
	GetEnvironmentById(context.Context, uuid.UUID) (repository.Environment, error)
}

// EnvironmentService implements environment business logic.
type EnvironmentService struct {
	repository EnvironmentRepository
}

// NewEnvironmentService builds an EnvironmentService backed by repo.
func NewEnvironmentService(repo EnvironmentRepository) *EnvironmentService {
	return &EnvironmentService{
		repository: repo,
	}
}

// CreateEnvironmentParams is the input to CreateEnvironment.
type CreateEnvironmentParams struct {
	Name string
	Slug string
}

// DeleteEnvironmentParams is the input to DeleteEnvironment.
type DeleteEnvironmentParams struct {
	Id uuid.UUID
}

// CreateEnvironment creates a new environment.
func (s *EnvironmentService) CreateEnvironment(ctx context.Context, arg *CreateEnvironmentParams) (*domain.Environment, error) {
	currentTime := time.Now().UTC()
	createEnvironmentParams := repository.CreateEnvironmentParams{
		Name:      arg.Name,
		Slug:      arg.Slug,
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
	environment, err := s.repository.CreateEnvironment(ctx, createEnvironmentParams)
	if err != nil {
		return nil, fmt.Errorf("create environment: %w", repository.MapError(err))
	}
	return toEnvironment(&environment), nil

}

// DeleteEnvironment deletes the environment identified by arg.Id.
func (s *EnvironmentService) DeleteEnvironment(ctx context.Context, arg *DeleteEnvironmentParams) error {
	_, err := s.repository.DeleteEnvironment(ctx, arg.Id)
	return repository.MapError(err)
}

// ListEnvironments returns every environment.
func (s *EnvironmentService) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	envs, err := s.repository.ListEnvironments(ctx)

	var environments = []domain.Environment{}
	if err != nil {
		return environments, fmt.Errorf("list environments:-%w", repository.MapError(err))
	}
	for _, env := range envs {
		environments = append(environments, *toEnvironment(&env))
	}
	return environments, nil

}

// GetEnvironmentByID returns the environment identified by environmentId.
func (s *EnvironmentService) GetEnvironmentByID(ctx context.Context, environmentID uuid.UUID) (*domain.Environment, error) {
	environment, err := s.repository.GetEnvironmentById(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("get environment: %w", repository.MapError(err))
	}
	return toEnvironment(&environment), nil
}

// toEnvironment converts a repository row into the domain type.
func toEnvironment(repositoryEnv *repository.Environment) *domain.Environment {
	return &domain.Environment{
		ID:        repositoryEnv.ID,
		Name:      repositoryEnv.Name,
		Slug:      repositoryEnv.Slug,
		CreatedAt: repositoryEnv.CreatedAt,
		UpdatedAt: repositoryEnv.UpdatedAt,
	}

}
