package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devpablocristo/argos/core/internal/config"
	"github.com/devpablocristo/argos/core/wire"
)

func main() {
	cfg := config.Load()
	handler, cleanup, err := wire.NewServer(wire.Config{
		DatabaseURL:          cfg.DatabaseURL,
		StorageDir:           cfg.StorageDir,
		ProcessingPython:     cfg.ProcessingPython,
		ProcessingPythonPath: cfg.ProcessingPythonPath,
		ProcessingTimeoutSec: int(cfg.ProcessingTimeout.Seconds()),
		OrgID:                cfg.OrgID,
		NexusBaseURL:         cfg.NexusBaseURL,
		NexusAPIKey:          cfg.NexusAPIKey,
		CompanionBaseURL:     cfg.CompanionBaseURL,
		CompanionAPIKey:      cfg.CompanionAPIKey,
		PublicBaseURL:        cfg.PublicBaseURL,
	})
	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("argos core listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
}
