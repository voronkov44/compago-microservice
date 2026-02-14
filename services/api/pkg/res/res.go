package res

import (
	"encoding/json"
	"net/http"
)

func Json(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, msg string, statusCode int) {
	Json(w, map[string]any{"error": msg}, statusCode)
}
