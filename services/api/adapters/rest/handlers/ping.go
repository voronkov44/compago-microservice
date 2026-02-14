package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"task-manager-microservice/api/core"
	"task-manager-microservice/api/pkg/res"
)

func NewPingHandler(log *slog.Logger, pingmap map[string]core.Pinger, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		out := map[string]string{}
		code := http.StatusOK

		for name, p := range pingmap {
			if err := p.Ping(ctx); err != nil {
				log.Warn("ping failed", "service", name, "error", err)
				out[name] = "down"
				code = http.StatusServiceUnavailable
			} else {
				out[name] = "ok"
			}
		}

		res.Json(w, map[string]any{"services": out}, code)
	}
}
