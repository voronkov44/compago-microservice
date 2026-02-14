package core

import "time"

type TaskStatus string

const (
	StatusTODO       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusArchived   TaskStatus = "archived"
)

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID          int64      `json:"id"`
	CategoryID  *int64     `json:"category_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskPatch struct {
	CategoryID  *int64
	Name        *string
	Description *string
	Status      *TaskStatus
}

type ListTasksFilter struct {
	Status          *TaskStatus
	CategoryID      *int64
	WithoutCategory bool
	Limit           int
	Offset          int
}
