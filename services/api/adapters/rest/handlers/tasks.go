package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"task-manager-microservice/api/adapters/rest"
	"time"

	"task-manager-microservice/api/core"
	"task-manager-microservice/api/pkg/res"
)

func parseStatus(s string) (core.TaskStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "todo":
		return core.StatusTODO, true
	case "in_progress":
		return core.StatusInProgress, true
	case "done":
		return core.StatusDone, true
	case "archived":
		return core.StatusArchived, true
	default:
		return "", false
	}
}

func NewCreateTaskHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in rest.CreateTaskIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			res.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		// category_id: nil - не передано, 0 - без категории
		var cid *int64
		if in.CategoryID != nil && *in.CategoryID != 0 {
			cid = in.CategoryID
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		t, err := svc.CreateTask(ctx, cid, in.Name, in.Description)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, t, http.StatusCreated)
	}
}

func NewGetTaskHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		t, err := svc.GetTask(ctx, id)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, t, http.StatusOK)
	}
}

func NewListTasksHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var f core.ListTasksFilter

		if s := q.Get("status"); s != "" {
			st, ok := parseStatus(s)
			if !ok {
				res.Error(w, "invalid status", http.StatusBadRequest)
				return
			}
			f.Status = &st
		}

		if v := q.Get("category_id"); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil || id <= 0 {
				res.Error(w, "invalid category_id", http.StatusBadRequest)
				return
			}
			f.CategoryID = &id
		}

		if v := q.Get("without_category"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				res.Error(w, "invalid without_category", http.StatusBadRequest)
				return
			}
			f.WithoutCategory = b
		}

		if f.CategoryID != nil && f.WithoutCategory {
			res.Error(w, "category_id and without_category are mutually exclusive", http.StatusBadRequest)
			return
		}

		if v := q.Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				res.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			f.Limit = n
		}
		if v := q.Get("offset"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				res.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}
			f.Offset = n
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		items, err := svc.ListTasks(ctx, f)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, map[string]any{"tasks": items}, http.StatusOK)
	}
}

func NewPatchTaskHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var in rest.PatchTaskIn
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			res.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		var p core.TaskPatch

		if in.CategoryID != nil {
			// 0 => снять категорию
			p.CategoryID = in.CategoryID
		}
		if in.Name != nil {
			p.Name = in.Name
		}
		if in.Description != nil {
			p.Description = in.Description
		}
		if in.Status != nil {
			st, ok := parseStatus(*in.Status)
			if !ok {
				res.Error(w, "invalid status", http.StatusBadRequest)
				return
			}
			p.Status = &st
		}

		if p.CategoryID == nil && p.Name == nil && p.Description == nil && p.Status == nil {
			res.Error(w, "no fields to update", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		t, err := svc.PatchTask(ctx, id, p)
		if err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, t, http.StatusOK)
	}
}

func NewDeleteTaskHandler(_ *slog.Logger, svc core.Tasks, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if err := svc.DeleteTask(ctx, id); err != nil {
			rest.WriteErr(w, err)
			return
		}
		res.Json(w, map[string]any{"ok": true}, http.StatusOK)
	}
}
