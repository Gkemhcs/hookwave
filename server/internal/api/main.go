package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/observability"
)

type HttpServer struct {
	*http.Server
}

func NewServer(logger *observability.Logger, port int) *HttpServer {
	router := NewRouter(logger)
	server := HttpServer{
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
