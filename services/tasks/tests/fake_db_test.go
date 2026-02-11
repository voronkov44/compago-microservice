package tests

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"task-manager-microservice/tasks/core"
)

type fakeDB struct {
	mu sync.RWMutex

	nextCategoryID int64
	nextTaskID     int64

	categories map[int64]core.Category
	tasks      map[int64]core.Task
}

func newFakeDB() *fakeDB {
	return &fakeDB{
		nextCategoryID: 1,
		nextTaskID:     1,
		categories:     make(map[int64]core.Category),
		tasks:          make(map[int64]core.Task),
	}
}

func cloneTask(t core.Task) core.Task {
	out := t
	if t.CategoryID != nil {
		cid := *t.CategoryID
		out.CategoryID = &cid
	}
	return out
}

func (db *fakeDB) Ping(context.Context) error {
	return nil
}

func (db *fakeDB) CreateCategory(_ context.Context, name string) (core.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return core.Category{}, core.ErrCategoryInvalidArgs
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	id := db.nextCategoryID
	db.nextCategoryID++

	category := core.Category{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
	db.categories[id] = category

	return category, nil
}

func (db *fakeDB) GetCategory(_ context.Context, id int64) (core.Category, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	category, ok := db.categories[id]
	if !ok {
		return core.Category{}, core.ErrCategoryNotFound
	}
	return category, nil
}

func (db *fakeDB) ListCategories(context.Context) ([]core.Category, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	out := make([]core.Category, 0, len(db.categories))
	for _, category := range db.categories {
		out = append(out, category)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out, nil
}

func (db *fakeDB) UpdateCategory(_ context.Context, id int64, name string) (core.Category, error) {
	name = strings.TrimSpace(name)
	if id <= 0 || name == "" {
		return core.Category{}, core.ErrCategoryInvalidArgs
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	category, ok := db.categories[id]
	if !ok {
		return core.Category{}, core.ErrCategoryNotFound
	}

	category.Name = name
	db.categories[id] = category

	return category, nil
}

func (db *fakeDB) DeleteCategory(_ context.Context, id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.categories[id]; !ok {
		return core.ErrCategoryNotFound
	}

	delete(db.categories, id)

	for taskID, task := range db.tasks {
		if task.CategoryID != nil && *task.CategoryID == id {
			task.CategoryID = nil
			db.tasks[taskID] = task
		}
	}

	return nil
}

func (db *fakeDB) CreateTask(_ context.Context, categoryID *int64, name, description string) (core.Task, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return core.Task{}, core.ErrTaskInvalidArgs
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if categoryID != nil {
		if *categoryID <= 0 {
			return core.Task{}, core.ErrTaskInvalidArgs
		}
		if _, ok := db.categories[*categoryID]; !ok {
			return core.Task{}, core.ErrCategoryNotFound
		}
	}

	id := db.nextTaskID
	db.nextTaskID++

	now := time.Now()
	task := core.Task{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      core.TODO,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if categoryID != nil {
		cid := *categoryID
		task.CategoryID = &cid
	}

	db.tasks[id] = cloneTask(task)
	return cloneTask(task), nil
}

func (db *fakeDB) GetTask(_ context.Context, id int64) (core.Task, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	task, ok := db.tasks[id]
	if !ok {
		return core.Task{}, core.ErrTaskNotFound
	}
	return cloneTask(task), nil
}

func (db *fakeDB) ListTasks(_ context.Context, f core.ListTasksFilter) ([]core.Task, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	out := make([]core.Task, 0, len(db.tasks))
	for _, task := range db.tasks {
		if f.Status != nil && task.Status != *f.Status {
			continue
		}
		if f.CategoryID != nil {
			if task.CategoryID == nil || *task.CategoryID != *f.CategoryID {
				continue
			}
		}
		if f.WithoutCategory && task.CategoryID != nil {
			continue
		}
		out = append(out, cloneTask(task))
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	if f.Offset > len(out) {
		return []core.Task{}, nil
	}

	if f.Offset > 0 {
		out = out[f.Offset:]
	}

	if f.Limit > 0 && f.Limit < len(out) {
		out = out[:f.Limit]
	}

	return out, nil
}

func (db *fakeDB) UpdateTask(_ context.Context, t core.Task) (core.Task, error) {
	if t.ID <= 0 || strings.TrimSpace(t.Name) == "" {
		return core.Task{}, core.ErrTaskInvalidArgs
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	current, ok := db.tasks[t.ID]
	if !ok {
		return core.Task{}, core.ErrTaskNotFound
	}

	if t.CategoryID != nil {
		if *t.CategoryID <= 0 {
			return core.Task{}, core.ErrTaskInvalidArgs
		}
		if _, ok := db.categories[*t.CategoryID]; !ok {
			return core.Task{}, core.ErrCategoryNotFound
		}
	}

	t.Name = strings.TrimSpace(t.Name)
	t.CreatedAt = current.CreatedAt
	t.UpdatedAt = time.Now()

	db.tasks[t.ID] = cloneTask(t)
	return cloneTask(t), nil
}

func (db *fakeDB) DeleteTask(_ context.Context, id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.tasks[id]; !ok {
		return core.ErrTaskNotFound
	}

	delete(db.tasks, id)
	return nil
}
