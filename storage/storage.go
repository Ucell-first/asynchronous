// storage/task_storage.go
package storage

import (
	models "asynchronous/model/db"
	"context"
)

type IStorage interface {
	Task() ITaskStorage
	User() IUserStorage
	TaskResult() ITaskResultStorage
	Close()
}

type ITaskStorage interface {
	CreateTask(ctx context.Context, task models.Task) (string, error)
	GetTask(ctx context.Context, id string) (models.Task, error)
	UpdateTask(ctx context.Context, task models.Task) error
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status string) error
}

type ITaskResultStorage interface {
	CreateResult(ctx context.Context, result models.TaskResult) error
	GetResult(ctx context.Context, id string) (models.TaskResult, error)
	UpdateResult(ctx context.Context, result models.TaskResult) error
	DeleteResult(ctx context.Context, id string) error
	ListResultsByTask(ctx context.Context, taskID string) ([]models.TaskResult, error)
}

type IUserStorage interface {
	CreateUser(ctx context.Context, user models.User) (string, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
}
