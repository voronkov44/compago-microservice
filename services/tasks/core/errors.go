package core

import "errors"

// Categories errors
var (
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryInvalidArgs   = errors.New("category invalid args")
)

// Tasks errors
var (
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskInvalidArgs   = errors.New("task invalid args")
)
