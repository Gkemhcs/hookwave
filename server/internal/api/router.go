package api

import (
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/api/handlers"
	"github.com/Gkemhcs/hookwave/server/internal/api/middleware"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Health is a trivial liveness endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hookwave server is up🏋"))
}

// setupEnvironmentRoutes registers the environment CRUD routes.
func setupEnvironmentRoutes(r *chi.Mux, handler *handlers.EnvironmentHandler) {
	r.Route("/environments", func(r chi.Router) {
		r.Post("/", handler.CreateEnvironment)
		r.Get("/", handler.ListEnvironments)
		r.Delete("/{id}", handler.DeleteEnvironment)
		r.Get("/{id}", handler.GetEnvironmentById)
	})
}

// setupApplicationRoutes registers the application CRUD routes, nested
// under their owning environment.
func setupApplicationRoutes(r *chi.Mux, handler *handlers.ApplicationHandler) {
	r.Route("/environments/{environment_id}/applications", func(r chi.Router) {
		r.Post("/", handler.CreateApplication)
		r.Get("/", handler.GetAllApplicationsByEnvironmentID)
		r.Get("/{application_id}", handler.GetApplicationByID)
		r.Patch("/{application_id}", handler.UpdateApplicationByID)
		r.Delete("/{application_id}", handler.DeleteApplicationByID)
	})
}

// setupEndpointRoutes registers the endpoint CRUD routes plus the
// enable/disable actions, nested under their owning application.
func setupEndpointRoutes(r *chi.Mux, handler *handlers.EndpointHandler) {
	r.Route("/applications/{application_id}/endpoints", func(r chi.Router) {
		r.Post("/", handler.CreateEndpoint)
		r.Get("/", handler.GetAllEndpointsByApplicationID)
		r.Get("/{endpoint_id}", handler.GetEndpointByID)
		r.Patch("/{endpoint_id}", handler.UpdateEndpointByID)
		r.Delete("/{endpoint_id}", handler.DeleteEndpointByID)
		r.Patch("/{endpoint_id}/enable", handler.EnableEndpointByID)
		r.Patch("/{endpoint_id}/disable", handler.DisableEndpointByID)
	})
}

// setupMessageRoutes registers the endpoint CRUD routes plus the

func setupMessageRoutes(r *chi.Mux, handler *handlers.MessageHandler) {
	r.Route("/applications/{application_id}/messages", func(r chi.Router) {
		r.Post("/send", handler.Send)
	})
}

// NewRouter builds the chi router: request-ID/logging/recovery middleware,
// then every entity's routes.
func NewRouter(logger *logging.Logger, handlersRegistry *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(chiMiddleware.ClientIPFromRemoteAddr) // pick one ClientIPFrom* based on your infra, see below
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.InjectLogger(logger))
	r.Use(middleware.Logger)
	r.Get("/", Health)
	setupEnvironmentRoutes(r, handlersRegistry.Environment)
	setupApplicationRoutes(r, handlersRegistry.Application)
	setupEndpointRoutes(r, handlersRegistry.Endpoint)
	setupMessageRoutes(r,handlersRegistry.Message)
	return r

}
