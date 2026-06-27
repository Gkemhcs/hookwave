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
	"github.com/Gkemhcs/hookwave/server/internal/observability"
)

func main() {
	serverConfig := config.NewServerConfig()
	err := serverConfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config loading failed: %v\n", err)
		os.Exit(1)
	}

	logger := observability.NewLogger()
	defer logger.Sync()

	dbClient, err := db.NewHookWaveDBClient(serverConfig.DBConn)

	if err != nil {
		logger.Fatal("db connection failed", observability.NewTagError(err))
	}
	err = dbClient.Ping()
	if err != nil {
		logger.Fatal("db ping failed", observability.NewTagError(err))
	} else {
		logger.Info("successfully connected to database")
	}
	defer dbClient.Close()

	logger.Info("hookwave starting", observability.NewTag("port", serverConfig.Port))

	server := api.NewServer(logger, serverConfig.Port)

	go func() {
		logger.Info("starting server")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Fatal("server failed", observability.NewTagError(err))
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
			observability.NewTagError(err),
		)
	}

	logger.Info("server stopped")
}
