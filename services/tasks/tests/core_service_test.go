package tests

import (
	"context"
	"errors"
	"testing"

	"task-manager-microservice/tasks/core"
)

func newServiceWithFakeDB() (*fakeDB, *core.Service) {
	db := newFakeDB()
	return db, core.NewService(db)
}

func mustCreateCategory(t *testing.T, db *fakeDB, name string) core.Category {
	t.Helper()

	category, err := db.CreateCategory(context.Background(), name)
	if err != nil {
		t.Fatalf("failed to prepare category: %v", err)
	}
	return category
}

func mustCreateTask(t *testing.T, db *fakeDB, categoryID *int64, name, description string) core.Task {
	t.Helper()

	task, err := db.CreateTask(context.Background(), categoryID, name, description)
	if err != nil {
		t.Fatalf("failed to prepare task: %v", err)
	}
	return task
}

func TestServiceCreateTask_EmptyName(t *testing.T) {
	t.Parallel()

	_, svc := newServiceWithFakeDB()

	_, err := svc.CreateTask(context.Background(), nil, "   ", "description")
	if !errors.Is(err, core.ErrTaskInvalidArgs) {
		t.Fatalf("expected ErrTaskInvalidArgs, got %v", err)
	}
}

func TestServiceCreateTask_CategoryNotFound(t *testing.T) {
	t.Parallel()

	_, svc := newServiceWithFakeDB()

	missingCategoryID := int64(999)
	_, err := svc.CreateTask(context.Background(), &missingCategoryID, "task", "description")
	if !errors.Is(err, core.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestServiceCreateTask_CategoryExists_Success(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	category := mustCreateCategory(t, db, "work")
	categoryID := category.ID

	task, err := svc.CreateTask(context.Background(), &categoryID, "task", "description")
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}

	if task.CategoryID == nil {
		t.Fatalf("expected category id to be set")
	}
	if *task.CategoryID != categoryID {
		t.Fatalf("expected category id %d, got %d", categoryID, *task.CategoryID)
	}
}

func TestServicePatchTask_EmptyPatch(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	task := mustCreateTask(t, db, nil, "task", "description")

	var err error
	_, err = svc.PatchTask(context.Background(), task.ID, core.TaskPatch{})
	if !errors.Is(err, core.ErrTaskInvalidArgs) {
		t.Fatalf("expected ErrTaskInvalidArgs, got %v", err)
	}
}

func TestServicePatchTask_UpdateMaskLikeBehavior_NameOnlyDoesNotChangeOtherFields(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	category := mustCreateCategory(t, db, "work")
	categoryID := category.ID
	task := mustCreateTask(t, db, &categoryID, "old name", "old description")

	task.Status = core.Done
	task, err := db.UpdateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("failed to prepare task status: %v", err)
	}

	newName := "new name"
	updated, err := svc.PatchTask(context.Background(), task.ID, core.TaskPatch{Name: &newName})
	if err != nil {
		t.Fatalf("PatchTask returned error: %v", err)
	}

	if updated.Name != newName {
		t.Fatalf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Description != task.Description {
		t.Fatalf("expected description %q, got %q", task.Description, updated.Description)
	}
	if updated.Status != task.Status {
		t.Fatalf("expected status %v, got %v", task.Status, updated.Status)
	}
	if updated.CategoryID == nil {
		t.Fatalf("expected category to stay set")
	}
	if task.CategoryID == nil || *updated.CategoryID != *task.CategoryID {
		t.Fatalf("expected category id %v, got %v", task.CategoryID, updated.CategoryID)
	}
}

func TestServicePatchTask_DescriptionEmptyClearsDescription(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()
	task := mustCreateTask(t, db, nil, "task", "non-empty")

	emptyDescription := ""
	updated, err := svc.PatchTask(context.Background(), task.ID, core.TaskPatch{Description: &emptyDescription})
	if err != nil {
		t.Fatalf("PatchTask returned error: %v", err)
	}

	if updated.Description != "" {
		t.Fatalf("expected empty description, got %q", updated.Description)
	}
}

func TestServicePatchTask_CategoryIDZeroRemovesCategory(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	category := mustCreateCategory(t, db, "work")
	categoryID := category.ID
	task := mustCreateTask(t, db, &categoryID, "task", "description")

	zero := int64(0)
	updated, err := svc.PatchTask(context.Background(), task.ID, core.TaskPatch{CategoryID: &zero})
	if err != nil {
		t.Fatalf("PatchTask returned error: %v", err)
	}

	if updated.CategoryID != nil {
		t.Fatalf("expected category to be removed, got %v", *updated.CategoryID)
	}
}

func TestServicePatchTask_CategoryNotFound(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	task := mustCreateTask(t, db, nil, "task", "description")

	missingCategoryID := int64(999)
	var err error
	_, err = svc.PatchTask(context.Background(), task.ID, core.TaskPatch{CategoryID: &missingCategoryID})
	if !errors.Is(err, core.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestServicePatchTask_TaskNotFound(t *testing.T) {
	t.Parallel()

	_, svc := newServiceWithFakeDB()

	newName := "updated"
	_, err := svc.PatchTask(context.Background(), 999, core.TaskPatch{Name: &newName})
	if !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestServicePatchTask_InvalidStatus(t *testing.T) {
	t.Parallel()

	db, svc := newServiceWithFakeDB()

	task := mustCreateTask(t, db, nil, "task", "description")

	invalidStatus := core.TaskStatus(99)
	var err error
	_, err = svc.PatchTask(context.Background(), task.ID, core.TaskPatch{Status: &invalidStatus})
	if !errors.Is(err, core.ErrTaskInvalidArgs) {
		t.Fatalf("expected ErrTaskInvalidArgs, got %v", err)
	}
}
