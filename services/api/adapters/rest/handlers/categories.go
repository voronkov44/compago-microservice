package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"task-manager-microservice/api/adapters/rest"
	"time"

	"task-manager-microservice/api/core"
	"task-manager-microservice/api/pkg/res"
)

func NewCreateCategoryHandler(log *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in rest.CreateCategoryIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			res.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		c, err := svc.CreateCategory(ctx, in.Name)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, c, http.StatusCreated)
	}
}

func NewGetCategoryHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		c, err := svc.GetCategory(ctx, id)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, c, http.StatusOK)
	}
}

func NewListCategoriesHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		items, err := svc.ListCategories(ctx)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, map[string]any{"categories": items}, http.StatusOK)
	}
}

func NewUpdateCategoryHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var in rest.UpdateCategoryIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			res.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		c, err := svc.UpdateCategory(ctx, id, in.Name)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, c, http.StatusOK)
	}
}

func NewDeleteCategoryHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if err := svc.DeleteCategory(ctx, id); err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, map[string]any{"ok": true}, http.StatusOK)
	}
}
