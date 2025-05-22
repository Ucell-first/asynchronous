package service

import (
	models "asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type TaskService struct {
	Storage   storage.IStorage
	Logger    *slog.Logger
	TaskQueue chan models.Task // Buffered channel
}

func NewTaskService(db storage.IStorage, logger *slog.Logger) *TaskService {
	ts := &TaskService{
		Storage:   db,
		Logger:    logger,
		TaskQueue: make(chan models.Task, 100),
	}

	go ts.startWorker(3)
	return ts
}

func (s *TaskService) CreateTask(ctx context.Context, req models.Task) (string, error) {
	// Payloadni JSON ga o'tkazamiz
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return "", fmt.Errorf("payloadni marshal qilishda xato: %w", err)
	}

	req.ID = uuid.New().String()
	req.Status = "pending"
	req.Payload = payloadBytes // Endi []byte formatda
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if _, err := s.Storage.Task().CreateTask(ctx, req); err != nil {
		return "", err
	}

	s.TaskQueue <- req
	return req.ID, nil
}

func (s *TaskService) startWorker(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go s.taskWorker(i + 1)
	}
}

func (s *TaskService) taskWorker(workerID int) {
	for task := range s.TaskQueue {
		s.processTask(task, workerID)
	}
}

func (s *TaskService) processTask(task models.Task, workerID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Statusni yangilash
	if err := s.Storage.Task().UpdateTaskStatus(ctx, task.ID, "processing"); err != nil {
		s.Logger.Error("status yangilashda xato", "error", err)
		return
	}

	// Payloadni parse qilish
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		s.handleTaskFailure(ctx, task, "invalid_payload")
		return
	}

	// Biznes logika
	result, err := s.executeTaskLogic(payload)
	if err != nil {
		s.handleTaskFailure(ctx, task, err.Error())
		return
	}

	// Natijani saqlash
	if err := s.Storage.TaskResult().CreateResult(ctx, models.TaskResult{
		ID:      uuid.New().String(),
		TaskID:  task.ID,
		FileURL: result.FileURL,
		GitURL:  result.GitURL,
	}); err != nil {
		s.Logger.Error("natija saqlashda xato", "error", err)
	}

	// Yakuniy status
	if err := s.Storage.Task().UpdateTaskStatus(ctx, task.ID, "completed"); err != nil {
		s.Logger.Error("status yangilashda xato", "error", err)
	}
}

// Qolgan metodlar o'zgarishsiz...

// Xato holatlarini boshqarish
func (s *TaskService) handleTaskFailure(ctx context.Context, task models.Task, errorMsg string) {
	// Qayta urinishlar sonini yangilash
	currentRetries := task.Retries + 1
	if currentRetries >= task.MaxRetries {
		s.Storage.Task().UpdateTaskStatus(ctx, task.ID, "failed")
		s.Logger.Error("Task maksimal qayta urinishlar soniga yetdi",
			"task_id", task.ID,
			"max_retries", task.MaxRetries,
		)
		return
	}

	// Exponential backoff
	retryDelay := time.Duration(2^currentRetries) * time.Second
	s.Logger.Info("Task qayta uriniladi",
		"task_id", task.ID,
		"retry", currentRetries,
		"delay", retryDelay,
	)

	time.AfterFunc(retryDelay, func() {
		s.TaskQueue <- task
	})
}

// Biznes logika (Sizning mantiqingizga moslashtiring)
func (s *TaskService) executeTaskLogic(payload map[string]interface{}) (*models.TaskResult, error) {
	// Misol: Faylni S3 ga yuklash
	file, ok := payload["file"].(string)
	if !ok {
		return nil, fmt.Errorf("payloadda file maydoni topilmadi")
	}

	// Bu yerda haqiqiy yuklash logikangiz bo'ladi
	result := &models.TaskResult{
		FileURL: "https://s3.example.com/" + file,
		GitURL:  "https://github.com/project/commit",
	}

	return result, nil
}
