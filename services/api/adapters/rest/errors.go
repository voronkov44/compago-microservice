package rest

import (
	"errors"
	"net/http"

	"task-manager-microservice/api/core"
	"task-manager-microservice/api/pkg/res"
)

func WriteErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrBadArguments):
		res.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, core.ErrNotFound):
		res.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, core.ErrAlreadyExists):
		res.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, core.ErrUnavailable):
		res.Error(w, err.Error(), http.StatusServiceUnavailable)
	default:
		res.Error(w, "internal error", http.StatusInternalServerError)
	}
}
