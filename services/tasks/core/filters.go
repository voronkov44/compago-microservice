package core

type ListTasksFilter struct {
	Status          *TaskStatus `json:"status"`
	CategoryID      *int64      `json:"category_id"`
	WithoutCategory bool        `json:"without_category"`
	Limit           int         `json:"limit"`
	Offset          int         `json:"offset"`
}
