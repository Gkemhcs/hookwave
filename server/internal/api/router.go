package api

import (
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/api/middleware"
	"github.com/Gkemhcs/hookwave/server/internal/observability"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("om namah sivaiah"))
}

func NewRouter(logger *observability.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.SetRequestHttpContext(logger))
	r.Use(middleware.Logger)
	r.Use(chiMiddleware.ClientIPFromRemoteAddr) // pick one ClientIPFrom* based on your infra, see below
	r.Use(chiMiddleware.Recoverer)
	r.Get("/", Health)
	return r

}
