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

type TaskPatch struct {
	CategoryID  *int64
	Name        *string
	Description *string
	Status      *TaskStatus
}

func (s *Service) CreateTask(ctx context.Context, categoryID *int64, name, description string) (Task, error) {
	if strings.TrimSpace(name) == "" {
		return Task{}, ErrTaskInvalidArgs
	}

	if categoryID != nil {
		if *categoryID <= 0 {
			return Task{}, ErrTaskInvalidArgs
		}
		if _, err := s.db.GetCategory(ctx, *categoryID); err != nil {
			return Task{}, err
		}
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

	if t.CategoryID != nil {
		if *t.CategoryID <= 0 {
			return Task{}, ErrTaskInvalidArgs
		}
		if _, err := s.db.GetCategory(ctx, *t.CategoryID); err != nil {
			return Task{}, err
		}
	}

	return s.db.UpdateTask(ctx, t)
}

func (s *Service) PatchTask(ctx context.Context, id int64, p TaskPatch) (Task, error) {
	if id <= 0 {
		return Task{}, ErrTaskInvalidArgs
	}
	if p.CategoryID == nil && p.Name == nil && p.Description == nil && p.Status == nil {
		return Task{}, ErrTaskInvalidArgs
	}

	cur, err := s.db.GetTask(ctx, id)
	if err != nil {
		return Task{}, err // ErrTaskNotFound -> NotFound
	}

	if p.Name != nil {
		name := strings.TrimSpace(*p.Name)
		if name == "" {
			return Task{}, ErrTaskInvalidArgs
		}
		cur.Name = name
	}

	if p.Description != nil {
		cur.Description = strings.TrimSpace(*p.Description)
	}

	if p.Status != nil {
		if !isValidStatus(*p.Status) {
			return Task{}, ErrTaskInvalidArgs
		}
		cur.Status = *p.Status
	}

	if p.CategoryID != nil {
		if *p.CategoryID < 0 {
			return Task{}, ErrTaskInvalidArgs
		}

		if *p.CategoryID == 0 {
			// remove category
			cur.CategoryID = nil
		} else {
			cid := *p.CategoryID
			// requirement: if category_id set -> check existence -> NotFound
			if _, err := s.db.GetCategory(ctx, cid); err != nil {
				return Task{}, err
			}
			cur.CategoryID = &cid
		}
	}

	return s.db.UpdateTask(ctx, cur)
}

func (s *Service) DeleteTask(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrTaskInvalidArgs
	}
	return s.db.DeleteTask(ctx, id)
}
