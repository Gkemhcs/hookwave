// Command server is Hookwave's main HTTP server entrypoint. It loads
// configuration, connects to Postgres, wires up the API server, and runs it
// until SIGINT/SIGTERM triggers a graceful shutdown.
package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/api"
	"github.com/Gkemhcs/hookwave/server/internal/config"
	"github.com/Gkemhcs/hookwave/server/internal/db"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging/tag"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
)

func main() {
	serverConfig := config.New()
	err := serverConfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config loading failed: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New()
	defer logger.Sync()

	hookwaveDBClient, err := db.New(serverConfig.DBConn)
	if err != nil {
		logger.Fatal("db connection failed", tag.NewTagError(err))
	}
	err = hookwaveDBClient.Ping()
	if err != nil {
		logger.Fatal("db ping failed", tag.NewTagError(err))
	} else {
		logger.Info("successfully connected to database")
	}
	defer hookwaveDBClient.Close()

	hookewaveRepository := repository.New(hookwaveDBClient)

	logger.Info("hookwave starting", tag.NewTag("port", serverConfig.Port))

	server := api.NewServer(logger, serverConfig.Port, hookewaveRepository)

	// Run the HTTP server in its own goroutine so the main goroutine stays
	// free to wait for a shutdown signal below.
	go func() {
		logger.Info("starting server")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Fatal("server failed", tag.NewTagError(err))
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed",
			tag.NewTagError(err),
		)
	}

	logger.Info("server stopped")
}
