// storage/postgres/task_result_repository.go
package postgres

import (
	models "asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskResultRepository struct {
	db *sql.DB
}

func NewTaskResultRepository(db *sql.DB) storage.ITaskResultStorage {
	return &TaskResultRepository{db: db}
}

func (r *TaskResultRepository) CreateResult(ctx context.Context, result models.TaskResult) error {
	result.ID = uuid.New().String()
	query := `
		INSERT INTO task_results (id, task_id, file_url, git_url, completed_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		result.ID,
		result.TaskID,
		result.FileURL,
		result.GitURL,
		time.Now(),
	)

	return err
}

func (r *TaskResultRepository) GetResult(ctx context.Context, id string) (models.TaskResult, error) {
	var result models.TaskResult
	query := `SELECT id, task_id, file_url, git_url, completed_at 
			  FROM task_results WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.TaskID,
		&result.FileURL,
		&result.GitURL,
		&result.CompletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.TaskResult{}, fmt.Errorf("natija topilmadi")
	}
	return result, err
}

func (r *TaskResultRepository) UpdateResult(ctx context.Context, result models.TaskResult) error {
	query := `
		UPDATE task_results SET
			file_url = $2,
			git_url = $3
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		result.ID,
		result.FileURL,
		result.GitURL,
	)

	return err
}

func (r *TaskResultRepository) DeleteResult(ctx context.Context, id string) error {
	query := `DELETE FROM task_results WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TaskResultRepository) ListResultsByTask(ctx context.Context, taskID string) ([]models.TaskResult, error) {
	query := `SELECT id, task_id, file_url, git_url, completed_at 
			  FROM task_results WHERE task_id = $1`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("natijalarni olishda xato: %w", err)
	}
	defer rows.Close()

	var results []models.TaskResult
	for rows.Next() {
		var result models.TaskResult
		if err := rows.Scan(
			&result.ID,
			&result.TaskID,
			&result.FileURL,
			&result.GitURL,
			&result.CompletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
