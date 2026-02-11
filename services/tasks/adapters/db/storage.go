package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"strings"
	"task-manager-microservice/tasks/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}
	return &DB{log: log, conn: db}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Categories

func (db *DB) CreateCategory(ctx context.Context, name string) (core.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return core.Category{}, core.ErrCategoryInvalidArgs
	}

	const q = `
		INSERT INTO categories(name)
		VALUES ($1)
		RETURNING id, created_at;
	`

	var c core.Category
	c.Name = name

	if err := db.conn.QueryRowxContext(ctx, q, c.Name).Scan(&c.ID, &c.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return core.Category{}, core.ErrCategoryAlreadyExists
		}
		return core.Category{}, fmt.Errorf("insert category: %w", err)
	}
	return c, nil
}

func (db *DB) GetCategory(ctx context.Context, id int64) (core.Category, error) {
	const q = `SELECT id, name, created_at FROM categories WHERE id = $1`

	var c core.Category
	if err := db.conn.GetContext(ctx, &c, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Category{}, core.ErrCategoryNotFound
		}
		return core.Category{}, fmt.Errorf("get category: %w", err)
	}
	return c, nil
}

func (db *DB) ListCategories(ctx context.Context) ([]core.Category, error) {
	const q = `SELECT id, name, created_at FROM categories ORDER BY lower(name) ASC`

	var out []core.Category
	if err := db.conn.SelectContext(ctx, &out, q); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return out, nil
}

func (db *DB) UpdateCategory(ctx context.Context, id int64, name string) (core.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return core.Category{}, core.ErrCategoryInvalidArgs
	}

	const q = `
		UPDATE categories
		SET name = $2
		WHERE id = $1
		RETURNING id, name, created_at;
	`

	var c core.Category
	if err := db.conn.GetContext(ctx, &c, q, id, name); err != nil {
		if isUniqueViolation(err) {
			return core.Category{}, core.ErrCategoryAlreadyExists
		}
		if errors.Is(err, sql.ErrNoRows) {
			return core.Category{}, core.ErrCategoryNotFound
		}
		return core.Category{}, fmt.Errorf("update category: %w", err)
	}
	return c, nil
}

func (db *DB) DeleteCategory(ctx context.Context, id int64) error {
	const q = `DELETE FROM categories WHERE id = $1`

	res, err := db.conn.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return core.ErrCategoryNotFound
	}
	return nil
}

// Tasks

func (db *DB) CreateTask(ctx context.Context, categoryID *int64, name, description string) (core.Task, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return core.Task{}, core.ErrTaskInvalidArgs
	}

	const q = `
		INSERT INTO tasks(category_id, name, description, status)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		RETURNING id, category_id, name, COALESCE(description, ''), status, created_at, updated_at;
	`

	status := core.TODO

	var t core.Task
	err := db.conn.QueryRowxContext(ctx, q, categoryID, name, strings.TrimSpace(description), int16(status)).
		Scan(&t.ID, &t.CategoryID, &t.Name, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		if isForeignKeyViolation(err) {
			return core.Task{}, core.ErrCategoryNotFound
		}
		if isCheckViolation(err) {
			return core.Task{}, core.ErrTaskInvalidArgs
		}
		return core.Task{}, fmt.Errorf("insert task: %w", err)
	}
	return t, nil
}

func (db *DB) GetTask(ctx context.Context, id int64) (core.Task, error) {
	const q = `
		SELECT id, category_id, name, COALESCE(description, '') AS description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1;
	`

	var t core.Task
	if err := db.conn.GetContext(ctx, &t, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.Task{}, core.ErrTaskNotFound
		}
		return core.Task{}, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (db *DB) ListTasks(ctx context.Context, f core.ListTasksFilter) ([]core.Task, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	var (
		sb   strings.Builder
		args []any
		n    = 1
	)

	sb.WriteString(`SELECT id, category_id, name, COALESCE(description, '') AS description, status, created_at, updated_at FROM tasks WHERE 1=1`)

	if f.Status != nil {
		args = append(args, int16(*f.Status))
		sb.WriteString(fmt.Sprintf(" AND status = $%d", n))
		n++
	}

	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		sb.WriteString(fmt.Sprintf(" AND category_id = $%d", n))
		n++
	} else if f.WithoutCategory {
		sb.WriteString(" AND category_id IS NULL")
	}

	args = append(args, f.Limit, f.Offset)
	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", n, n+1))

	var out []core.Task
	if err := db.conn.SelectContext(ctx, &out, sb.String(), args...); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return out, nil
}

func (db *DB) UpdateTask(ctx context.Context, t core.Task) (core.Task, error) {
	t.Name = strings.TrimSpace(t.Name)
	if t.ID == 0 || t.Name == "" {
		return core.Task{}, core.ErrTaskInvalidArgs
	}

	const q = `
		UPDATE tasks
		SET category_id = $2,
		    name = $3,
		    description = NULLIF($4, ''),
		    status = $5,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, category_id, name, COALESCE(description, '') AS description, status, created_at, updated_at;
	`

	var out core.Task
	if err := db.conn.GetContext(ctx, &out, q, t.ID, t.CategoryID, t.Name, strings.TrimSpace(t.Description), int16(t.Status)); err != nil {
		if isForeignKeyViolation(err) {
			return core.Task{}, core.ErrCategoryNotFound
		}
		if isCheckViolation(err) {
			return core.Task{}, core.ErrTaskInvalidArgs
		}
		if errors.Is(err, sql.ErrNoRows) {
			return core.Task{}, core.ErrTaskNotFound
		}
		return core.Task{}, fmt.Errorf("update task: %w", err)
	}
	return out, nil
}

func (db *DB) DeleteTask(ctx context.Context, id int64) error {
	const q = `DELETE FROM tasks WHERE id = $1`

	res, err := db.conn.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return core.ErrTaskNotFound
	}
	return nil
}

// pg helpers

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}
