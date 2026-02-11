package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	taskspb "task-manager-microservice/proto/tasks"
	"task-manager-microservice/tasks/adapters/db"
	taskgrpc "task-manager-microservice/tasks/adapters/grpc"
	"task-manager-microservice/tasks/config"
	"task-manager-microservice/tasks/core"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "tasks-service server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	// logger
	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting tasks-service server")

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// database adapter
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}
	// defer func.Close
	defer func(storage *db.DB) {
		err := storage.Close()
		if err != nil {
			log.Error("failed to close db connection", "error", err)
		}
	}(storage)

	if err := storage.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate db: %v", err)
	}

	// service
	tasksService := core.NewService(storage)

	// grpc
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// grpc handler
	handler := taskgrpc.NewServer(log, tasksService)

	taskspb.RegisterCategoriesServiceServer(s, handler)
	taskspb.RegisterTasksServiceServer(s, handler)
	reflection.Register(s)

	go func() {
		<-ctx.Done()
		log.Debug("shutting down tasks-service server")
		s.GracefulStop()
	}()

	log.Info("tasks-service gRPC server is running", "address", cfg.Address)

	// blocking
	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func mustMakeLogger(levelStr string) *slog.Logger {
	var level slog.Level
	switch levelStr {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
