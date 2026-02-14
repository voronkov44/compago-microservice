package rest

type CreateCategoryIn struct {
	Name string `json:"name"`
}

type UpdateCategoryIn struct {
	Name string `json:"name"`
}

type CreateTaskIn struct {
	CategoryID  *int64 `json:"category_id,omitempty"` // Nil - не передано, 0 - без категории
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PatchTaskIn struct {
	CategoryID  *int64  `json:"category_id,omitempty"` // Если передано 0 - снять категорию
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"` // todo|in_progress|done|archived
}
