package core

import "context"

type Pinger interface {
	Ping(ctx context.Context) error
}

type Tasks interface {
	Pinger

	// categories
	CreateCategory(ctx context.Context, name string) (Category, error)
	GetCategory(ctx context.Context, id int64) (Category, error)
	ListCategories(ctx context.Context) ([]Category, error)
	UpdateCategory(ctx context.Context, id int64, name string) (Category, error)
	DeleteCategory(ctx context.Context, id int64) error

	// tasks
	CreateTask(ctx context.Context, categoryID *int64, name, description string) (Task, error)
	GetTask(ctx context.Context, id int64) (Task, error)
	ListTasks(ctx context.Context, f ListTasksFilter) ([]Task, error)
	PatchTask(ctx context.Context, id int64, p TaskPatch) (Task, error)
	DeleteTask(ctx context.Context, id int64) error
}
