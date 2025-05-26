// models/task.go
package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Task struct {
	ID                  string          `json:"id"`
	CreatorID           string          `json:"creator_id"`
	UserID              string          `json:"user_id"`
	Title               string          `json:"title"`
	Priority            int             `json:"priority"`
	Status              string          `json:"status"`
	CanUserChangeStatus bool            `json:"can_user_change_status"`
	Payload             json.RawMessage `json:"payload"`
	Retries             int             `json:"retries"`
	MaxRetries          int             `json:"max_retries"`
	ScheduledAt         sql.NullTime    `json:"scheduled_at"`
	NextRetryAt         *time.Time      `json:"next_retry_at"` // Yangi qo'shilgan maydon
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	DeletedAt           sql.NullTime    `json:"deleted_at"`
}

type TaskResult struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	FileURL     string    `json:"file_url"`
	GitURL      string    `json:"git_url"`
	CompletedAt time.Time `json:"completed_at"`
}

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string
	Surname      string
	PasswordHash string       `json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at"`
}
