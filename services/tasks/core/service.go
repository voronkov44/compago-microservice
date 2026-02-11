package core

import (
	"context"
	"strings"
)

type Service struct {
	db DB
}

func NewService(db DB) *Service {
	return &Service{
		db: db,
	}
}

func isValidStatus(st TaskStatus) bool {
	return st >= TODO && st <= Archived
}

func (s *Service) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

// Categories

func (s *Service) CreateCategory(ctx context.Context, name string) (Category, error) {
	if strings.TrimSpace(name) == "" {
		return Category{}, ErrCategoryInvalidArgs
	}
	return s.db.CreateCategory(ctx, name)
}

func (s *Service) GetCategory(ctx context.Context, id int64) (Category, error) {
	if id <= 0 {
		return Category{}, ErrCategoryInvalidArgs
	}
	return s.db.GetCategory(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.db.ListCategories(ctx)
}

func (s *Service) UpdateCategory(ctx context.Context, id int64, name string) (Category, error) {
	if id <= 0 || strings.TrimSpace(name) == "" {
		return Category{}, ErrCategoryInvalidArgs
	}
	return s.db.UpdateCategory(ctx, id, name)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrCategoryInvalidArgs
	}
	return s.db.DeleteCategory(ctx, id)
}

// Tasks

func (s *Service) CreateTask(ctx context.Context, categoryID *int64, name, description string) (Task, error) {
	if strings.TrimSpace(name) == "" {
		return Task{}, ErrTaskInvalidArgs
	}
	if categoryID != nil && *categoryID <= 0 {
		categoryID = nil // 0 => без категории
	}
	return s.db.CreateTask(ctx, categoryID, name, description)
}

func (s *Service) GetTask(ctx context.Context, id int64) (Task, error) {
	if id <= 0 {
		return Task{}, ErrTaskInvalidArgs
	}
	return s.db.GetTask(ctx, id)
}

func (s *Service) ListTasks(ctx context.Context, f ListTasksFilter) ([]Task, error) {
	if f.Limit < 0 || f.Offset < 0 {
		return nil, ErrTaskInvalidArgs
	}
	if f.Status != nil && !isValidStatus(*f.Status) {
		return nil, ErrTaskInvalidArgs
	}
	if f.CategoryID != nil && *f.CategoryID <= 0 {
		return nil, ErrTaskInvalidArgs
	}
	if f.CategoryID != nil && f.WithoutCategory {
		return nil, ErrTaskInvalidArgs
	}
	return s.db.ListTasks(ctx, f)
}

func (s *Service) UpdateTask(ctx context.Context, t Task) (Task, error) {
	if t.ID <= 0 || strings.TrimSpace(t.Name) == "" {
		return Task{}, ErrTaskInvalidArgs
	}
	if !isValidStatus(t.Status) {
		return Task{}, ErrTaskInvalidArgs
	}
	if t.CategoryID != nil && *t.CategoryID <= 0 {
		t.CategoryID = nil // 0 => снять категорию
	}
	return s.db.UpdateTask(ctx, t)
}

func (s *Service) DeleteTask(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrTaskInvalidArgs
	}
	return s.db.DeleteTask(ctx, id)
}
