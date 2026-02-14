package handlers

import (
	"log/slog"
	"net/http"
	"task-manager-microservice/api/core"
	"time"
)

func Register(mux *http.ServeMux, log *slog.Logger, deps core.Deps, timeout time.Duration) {
	// ping
	mux.Handle("GET /api/ping", NewPingHandler(log, map[string]core.Pinger{"tasks": deps.Tasks}, timeout))

	// categories
	mux.Handle("POST /api/categories", NewCreateCategoryHandler(log, deps.Tasks, timeout))
	mux.Handle("GET /api/categories", NewListCategoriesHandler(log, deps.Tasks, timeout))
	mux.Handle("GET /api/categories/{id}", NewGetCategoryHandler(log, deps.Tasks, timeout))
	mux.Handle("PUT /api/categories/{id}", NewUpdateCategoryHandler(log, deps.Tasks, timeout))
	mux.Handle("DELETE /api/categories/{id}", NewDeleteCategoryHandler(log, deps.Tasks, timeout))

	// tasks
	mux.Handle("POST /api/tasks", NewCreateTaskHandler(log, deps.Tasks, timeout))
	mux.Handle("GET /api/tasks", NewListTasksHandler(log, deps.Tasks, timeout))
	mux.Handle("GET /api/tasks/{id}", NewGetTaskHandler(log, deps.Tasks, timeout))
	mux.Handle("PATCH /api/tasks/{id}", NewPatchTaskHandler(log, deps.Tasks, timeout))
	mux.Handle("DELETE /api/tasks/{id}", NewDeleteTaskHandler(log, deps.Tasks, timeout))
}
