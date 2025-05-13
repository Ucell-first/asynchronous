// storage/postgres/task_repository.go
package postgres

import (
	models "asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) storage.ITaskStorage {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task models.Task) (string, error) {
	task.ID = uuid.New().String()
	payload, _ := json.Marshal(task.Payload)

	query := `
		INSERT INTO tasks (
			id, creator_id, user_id, title, priority, status, 
			can_user_change_status, payload, retries, max_retries,
			scheduled_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.CreatorID,
		task.UserID,
		task.Title,
		task.Priority,
		task.Status,
		task.CanUserChangeStatus,
		payload,
		task.Retries,
		task.MaxRetries,
		task.ScheduledAt,
		time.Now(),
		time.Now(),
	)

	return task.ID, err
}

func (r *TaskRepository) GetTask(ctx context.Context, id string) (models.Task, error) {
	var task models.Task
	var payload []byte

	query := `
		SELECT 
			id, creator_id, user_id, title, priority, status, 
			can_user_change_status, payload, retries, max_retries,
			scheduled_at, created_at, updated_at, deleted_at
		FROM tasks 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.CreatorID,
		&task.UserID,
		&task.Title,
		&task.Priority,
		&task.Status,
		&task.CanUserChangeStatus,
		&payload,
		&task.Retries,
		&task.MaxRetries,
		&task.ScheduledAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Task{}, fmt.Errorf("task topilmadi")
	}
	if err != nil {
		return models.Task{}, fmt.Errorf("taskni olishda xato: %w", err)
	}

	json.Unmarshal(payload, &task.Payload)
	return task, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, task models.Task) error {
	payload, _ := json.Marshal(task.Payload)

	query := `
		UPDATE tasks SET
			title = $2,
			priority = $3,
			status = $4,
			can_user_change_status = $5,
			payload = $6,
			retries = $7,
			max_retries = $8,
			scheduled_at = $9,
			updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.Title,
		task.Priority,
		task.Status,
		task.CanUserChangeStatus,
		payload,
		task.Retries,
		task.MaxRetries,
		task.ScheduledAt,
		time.Now(),
	)

	return err
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id string) error {
	query := `UPDATE tasks SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *TaskRepository) ListTasks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.Task, error) {
	var tasks []models.Task
	var where []string
	args := []interface{}{}
	argIDx := 1

	baseQuery := `
		SELECT 
			id, creator_id, user_id, title, priority, status, 
			can_user_change_status, payload, retries, max_retries,
			scheduled_at, created_at, updated_at, deleted_at
		FROM tasks 
		WHERE deleted_at IS NULL`

	for key, val := range filters {
		switch key {
		case "creator_id":
			where = append(where, fmt.Sprintf("creator_id = $%d", argIDx))
			args = append(args, val)
			argIDx++
		case "user_id":
			where = append(where, fmt.Sprintf("user_id = $%d", argIDx))
			args = append(args, val)
			argIDx++
		case "status":
			where = append(where, fmt.Sprintf("status = $%d", argIDx))
			args = append(args, val)
			argIDx++
		}
	}

	if len(where) > 0 {
		baseQuery += " AND " + strings.Join(where, " AND ")
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIDx, argIDx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("tasklar ro'yxatini olishda xato: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		var payload []byte

		if err := rows.Scan(
			&task.ID,
			&task.CreatorID,
			&task.UserID,
			&task.Title,
			&task.Priority,
			&task.Status,
			&task.CanUserChangeStatus,
			&payload,
			&task.Retries,
			&task.MaxRetries,
			&task.ScheduledAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeletedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(payload, &task.Payload)
		tasks = append(tasks, task)
	}

	return tasks, nil
}
