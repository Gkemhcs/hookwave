// Package api is the composition root for the HTTP layer: it builds the
// services, wires them into handlers, builds the router, and assembles the
// http.Server.
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/api/handlers"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/Gkemhcs/hookwave/server/internal/service"
	"github.com/Gkemhcs/hookwave/server/internal/transactor"
)

// HTTPServer wraps an http.Server.
type HTTPServer struct {
	*http.Server
}

// NewServer builds every service on top of dbConn, wires them into
// handlers, and returns a ready-to-run HTTPServer listening on port.
func NewServer(logger *logging.Logger, port int, dbConn repository.Querier,transactor *transactor.Transactor) *HTTPServer {

	environmentService := service.NewEnvironmentService(dbConn)
	applicationService := service.NewApplicationService(dbConn)
	endpointService := service.NewEndpointService(dbConn)	
	messageService:=service.NewMessageService(transactor)
	
	handlers := handlers.NewHandlers(environmentService, applicationService, endpointService,messageService)
	router := NewRouter(logger, handlers)
	server := HTTPServer{
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           router,
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}}
	return &server

}
