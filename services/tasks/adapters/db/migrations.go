package db

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/01_create_categories.up.sql
var createCategoriesUp string

//go:embed migrations/02_create_tasks.up.sql
var createTasksUp string

// Migrate применяет миграции для task-сервиса
func (db *DB) Migrate() error {
	db.log.Debug("running tasksDB migrations")

	if _, err := db.conn.Exec(createCategoriesUp); err != nil {
		return fmt.Errorf("apply categories migration: %w", err)
	}

	if _, err := db.conn.Exec(createTasksUp); err != nil {
		return fmt.Errorf("apply tasks migration: %w", err)
	}

	db.log.Debug("tasksDB migrations finished")
	return nil
}
