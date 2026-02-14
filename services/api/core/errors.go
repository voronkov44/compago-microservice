package core

import "errors"

var (
	ErrBadArguments  = errors.New("bad arguments")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrUnavailable   = errors.New("dependency unavailable")
)
