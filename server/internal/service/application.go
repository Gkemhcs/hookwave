package service

import (
	"context"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/google/uuid"
)

// ApplicationRepository is the subset of the generated repository.Querier
// that ApplicationService depends on.
type ApplicationRepository interface {
	CreateApplication(ctx context.Context, arg repository.CreateApplicationParams) (repository.Application, error)
	DeleteApplicationByID(ctx context.Context, arg repository.DeleteApplicationByIDParams) error
	GetAllApplicationsByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]repository.Application, error)
	GetApplicationByID(ctx context.Context, id repository.GetApplicationByIDParams) (repository.Application, error)
	UpdateApplicationById(ctx context.Context, arg repository.UpdateApplicationByIdParams) (repository.Application, error)
}

// CreateApplicationParams is the input to CreateApplication.
type CreateApplicationParams struct {
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	EnvironmentID uuid.UUID `json:"environment_id"`
}

// GetApplicationByIDParams identifies one application by its environment
// and application IDs.
type GetApplicationByIDParams struct {
	EnvironmentID uuid.UUID `json:"environment_id"`
	ApplicationID uuid.UUID `json:"application_id"`
}

// DeleteApplicationByIDParams is the input to DeleteApplicationByID.
type DeleteApplicationByIDParams struct {
	ID            uuid.UUID `json:"environment_id"`
	ApplicationID uuid.UUID `json:"application_id"`
}

// UpdateApplicationByIDParams is the input to UpdateApplicationByID. An
// empty Name/Slug is treated as "leave this field unchanged".
type UpdateApplicationByIDParams struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	EnvironmentID uuid.UUID `json:"environment_id"`
}

// ApplicationService implements application business logic.
type ApplicationService struct {
	repository ApplicationRepository
}

// NewApplicationService builds an ApplicationService backed by repository.
func NewApplicationService(repository ApplicationRepository) *ApplicationService {
	return &ApplicationService{
		repository: repository,
	}
}

// CreateApplication creates a new application under an environment.
func (s *ApplicationService) CreateApplication(ctx context.Context, args *CreateApplicationParams) (*domain.Application, error) {
	currentTime := time.Now().UTC()
	applicationParams := repository.CreateApplicationParams{
		ID:            uuid.New(),
		Name:          args.Name,
		Slug:          args.Slug,
		EnvironmentID: args.EnvironmentID,
		CreatedAt:     currentTime,
		UpdatedAt:     currentTime,
	}
	application, err := s.repository.CreateApplication(ctx, applicationParams)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildApplication(&application), nil
}

// GetApplicationByID returns the application identified by args, scoped to
// its environment.
func (s *ApplicationService) GetApplicationByID(ctx context.Context, args *GetApplicationByIDParams) (*domain.Application, error) {
	application, err := s.repository.GetApplicationByID(ctx, repository.GetApplicationByIDParams{
		ID:            args.ApplicationID,
		EnvironmentID: args.EnvironmentID,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildApplication(&application), nil
}

// GetAllApplicationsByEnvironmentID returns every application under
// environmentID.
func (s *ApplicationService) GetAllApplicationsByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*domain.Application, error) {

	repositoryApplications, err := s.repository.GetAllApplicationsByEnvironmentID(ctx, environmentID)

	var applications []*domain.Application
	if err != nil {
		return applications, repository.MapError(err)
	}
	for _, application := range repositoryApplications {
		applications = append(applications, s.buildApplication(&application))
	}
	return applications, nil

}

// DeleteApplicationByID deletes the application identified by args, scoped
// to its environment.
func (s *ApplicationService) DeleteApplicationByID(ctx context.Context, args *DeleteApplicationByIDParams) error {
	err := s.repository.DeleteApplicationByID(ctx, repository.DeleteApplicationByIDParams{
		ID:            args.ApplicationID,
		EnvironmentID: args.ID,
	})
	if err != nil {
		return repository.MapError(err)
	}
	return nil
}

// UpdateApplicationByID applies a partial update to the application
// identified by args, scoped to its environment. An empty Name/Slug in
// args leaves the existing value unchanged.
func (s *ApplicationService) UpdateApplicationByID(ctx context.Context, args *UpdateApplicationByIDParams) (*domain.Application, error) {
	application, err := s.repository.GetApplicationByID(ctx, repository.GetApplicationByIDParams{
		ID:            args.ID,
		EnvironmentID: args.EnvironmentID,
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	updateApplicationParams := repository.UpdateApplicationByIdParams{
		Name:          application.Name,
		Slug:          application.Slug,
		ID:            application.ID,
		EnvironmentID: application.EnvironmentID,
	}
	if args.Name == "" {
		updateApplicationParams.Name = args.Name
	}
	if args.Slug == "" {
		updateApplicationParams.Slug = args.Slug
	}
	application, err = s.repository.UpdateApplicationById(ctx, updateApplicationParams)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildApplication(&application), nil
}

// buildApplication converts a repository row into the domain type.
func (s *ApplicationService) buildApplication(application *repository.Application) *domain.Application {
	return &domain.Application{
		ID:            application.ID,
		Name:          application.Name,
		Slug:          application.Slug,
		EnvironmentID: application.EnvironmentID,
		CreatedAt:     application.CreatedAt,
		UpdatedAt:     application.UpdatedAt,
	}
}
