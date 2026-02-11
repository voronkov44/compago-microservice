package core

import "time"

type TaskStatus int16

const (
	TODO       TaskStatus = 0
	InProgress TaskStatus = 1
	Done       TaskStatus = 2
	Archived   TaskStatus = 3
)

type Task struct {
	ID          int64      `db:"id"`
	CategoryID  *int64     `db:"category_id"` // Nil без категории
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Status      TaskStatus `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}
