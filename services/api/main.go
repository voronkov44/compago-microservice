package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"task-manager-microservice/api/adapters/rest/handlers"
	"task-manager-microservice/api/adapters/tasks"
	"task-manager-microservice/api/core"
	"time"

	"task-manager-microservice/api/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	log := mustMakeLogger(cfg.LogLevel)

	tasksClient, err := tasks.NewClient(cfg.TasksAddress, log)
	if err != nil {
		log.Error("cannot init tasks adapter", "error", err)
		os.Exit(1)
	}
	defer func() { _ = tasksClient.Close() }()

	deps := core.Deps{
		Tasks: tasksClient,
	}

	mux := http.NewServeMux()
	handlers.Register(mux, log, deps, cfg.HTTP.Timeout)

	server := http.Server{
		Addr:              cfg.HTTP.Address,
		ReadHeaderTimeout: cfg.HTTP.Timeout,
		Handler:           mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("api gateway http server", "address", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped unexpectedly", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
