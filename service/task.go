// service/task_service.go
package service

import (
	"asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
)

// TaskService - tasklarni boshqarish uchun asosiy service
type TaskService struct {
	storage    storage.IStorage
	logger     *slog.Logger
	taskQueue  chan *db.Task // Tasklar uchun kanal
	workerPool *WorkerPool   // Workerlar pooli
}

// NewTaskService - yangi TaskService yaratish
func NewTaskService(
	pdb storage.IStorage,
	logger *slog.Logger,
	workerCount int,
) *TaskService {
	// Kanalni initsializatsiya qilish (buffer size 1000)
	taskQueue := make(chan *db.Task, 1000)

	return &TaskService{
		storage:    pdb,
		logger:     logger,
		taskQueue:  taskQueue,
		workerPool: NewWorkerPool(pdb, logger, workerCount, taskQueue),
	}
}

// StartWorkers - workerlarni ishga tushirish
func (s *TaskService) StartWorkers() {
	s.workerPool.Start()
}

// CreateTask - yangi task yaratish va navbatga qo'shish
func (s *TaskService) CreateTask(ctx context.Context, req db.Task) (*db.Task, error) {
	// Validatsiyalar
	if req.Title == "" {
		return nil, errors.New("title bo'sh bo'lishi mumkin emas")
	}

	// Avtomatik to'ldirish
	req.ID = uuid.NewString()
	req.Status = "pending"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	// Bazaga saqlash
	if _, err := s.storage.Task().CreateTask(ctx, req); err != nil {
		return nil, fmt.Errorf("taskni saqlashda xato: %w", err)
	}

	// Navbatga pointer orqali qo'shish
	go func(task db.Task) {
		select {
		case s.taskQueue <- &task:
			s.logger.Info("Task navbatga qo'shildi", "task_id", task.ID)
		default:
			s.logger.Error("Navbat to'ldi, task qabul qilinmadi")
		}
	}(req)

	return &req, nil
}

// WorkerPool - tasklarni bajaruvchi ishchilar pooli
type WorkerPool struct {
	db          storage.IStorage
	logger      *slog.Logger
	workerCount int
	taskQueue   chan *db.Task
}

// NewWorkerPool - yangi WorkerPool yaratish
func NewWorkerPool(
	db storage.IStorage,
	logger *slog.Logger,
	workerCount int,
	taskQueue chan *db.Task, // Chanelni parameter sifatida qabul qilish
) *WorkerPool {
	return &WorkerPool{
		db:          db,
		logger:      logger,
		workerCount: workerCount,
		taskQueue:   taskQueue, // Kanalni saqlash
	}
}

// Start - workerlarni ishga tushirish
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		workerID := i + 1
		go wp.worker(workerID)
	}
}

// worker - har bir individual ishchi
func (wp *WorkerPool) worker(workerID int) {
	wp.logger.Info("Worker ishga tushdi", "worker_id", workerID)

	for task := range wp.taskQueue {
		wp.processTask(workerID, task)
	}
}

// processTask - taskni bajarish logikasi
func (wp *WorkerPool) processTask(workerID int, task *db.Task) {
	// ctx := context.Background()

	// 1. Statusni "processing" ga o'zgartirish
	if err := wp.updateTaskStatus(task, "processing"); err != nil {
		return
	}

	// 2. Taskni bajarish
	err := wp.executeTaskLogic(task)
	if err != nil {
		wp.handleTaskError(task, err)
		return
	}

	// 3. Muvaffaqiyatli yakunlash
	if err := wp.updateTaskStatus(task, "completed"); err != nil {
		return
	}

	// 4. Natijani saqlash
	if err := wp.saveTaskResult(task); err != nil {
		wp.logger.Error("Natijani saqlashda xato", "error", err)
	}
}

// executeTaskLogic - taskning asosiy logikasi
func (wp *WorkerPool) executeTaskLogic(task *db.Task) error {
	// Simulyatsiya: Taskni bajarish uchun vaqt
	if task.ScheduledAt.Valid && time.Now().Before(task.ScheduledAt.Time) {
		return fmt.Errorf("task hali bajarilish vaqti kelmagan")
	}

	// Payloadni parse qilish
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("payloadni parse qilishda xato: %w", err)
	}

	// Asosiy logika (sizning biznes mantiqingiz)
	wp.logger.Info("Task bajarilmoqda...", "task_id", task.ID)
	time.Sleep(1 * time.Second) // Simulyatsiya

	return nil
}

// handleTaskError - xatolikni boshqarish
func (wp *WorkerPool) handleTaskError(task *db.Task, err error) {
	wp.logger.Error("Taskda xato yuz berdi",
		"task_id", task.ID,
		"error", err.Error(),
	)

	// Qayta urinishlar sonini yangilash
	task.Retries++

	// Qayta urinishlar chegarasini tekshirish
	if task.Retries >= task.MaxRetries {
		wp.logger.Error("Maksimal qayta urinishlar soniga yetildi",
			"task_id", task.ID,
			"max_retries", task.MaxRetries,
		)
		_ = wp.updateTaskStatus(task, "failed")
		return
	}

	// Eksponensial kechikni hisoblash
	delay := time.Duration(math.Pow(2, float64(task.Retries))) * time.Second
	nextRetry := time.Now().Add(delay)
	task.NextRetryAt = &nextRetry

	// Taskni yangilash
	if err := wp.db.Task().UpdateTask(context.Background(), *task); err != nil {
		wp.logger.Error("Taskni yangilashda xato", "error", err)
		return
	}

	// Navbatga qayta qo'shish
	go func(t *db.Task) {
		time.Sleep(delay)
		wp.taskQueue <- t
	}(task)
}

// updateTaskStatus - task statusini yangilash
func (wp *WorkerPool) updateTaskStatus(task *db.Task, status string) error {
	task.Status = status
	task.UpdatedAt = time.Now()

	if err := wp.db.Task().UpdateTask(context.Background(), *task); err != nil {
		wp.logger.Error("Statusni yangilashda xato",
			"task_id", task.ID,
			"error", err.Error(),
		)
		return err
	}
	return nil
}

// saveTaskResult - task natijasini saqlash
func (wp *WorkerPool) saveTaskResult(task *db.Task) error {
	result := db.TaskResult{
		ID:          uuid.NewString(),
		TaskID:      task.ID,
		CompletedAt: time.Now(),
	}

	return wp.db.TaskResult().CreateResult(context.Background(), result)
}
