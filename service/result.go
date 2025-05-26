// service/result_service.go
package service

import (
	"asynchronous/model/db"
	"asynchronous/storage"
	"asynchronous/storage/postgres"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

type ResultService struct {
	storage storage.IStorage
	logger  *slog.Logger
}

func NewResultService(db *sql.DB, logger *slog.Logger) *ResultService {
	return &ResultService{
		storage: postgres.NewPostgresStorage(db),
		logger:  logger,
	}
}

// GetResult - Natijani ID bo'yicha olish
func (s *ResultService) GetResult(ctx context.Context, resultID string) (*db.TaskResult, error) {
	s.logger.Info("Natijani olish", "result_id", resultID)

	result, err := s.storage.TaskResult().GetResult(ctx, resultID)
	if err != nil {
		s.logger.Error("Natijani olishda xato", "error", err)
		return nil, fmt.Errorf("server ichki xatosi")
	}
	return &result, nil
}

// ListResultsByTask - Task uchun barcha natijalarni olish
func (s *ResultService) ListResultsByTask(ctx context.Context, taskID string) ([]db.TaskResult, error) {
	s.logger.Info("Task natijalarini olish", "task_id", taskID)

	results, err := s.storage.TaskResult().ListResultsByTask(ctx, taskID)
	if err != nil {
		s.logger.Error("Natijalarni olishda xato", "error", err)
		return nil, fmt.Errorf("natijalarni olishda xato")
	}

	if len(results) == 0 {
		s.logger.Warn("Task uchun natijalar topilmadi", "task_id", taskID)
		return nil, fmt.Errorf("natijalar topilmadi")
	}
	return results, nil
}

// CreateResult - Yangi natija yaratish (asosan workerlar uchun)
func (s *ResultService) CreateResult(ctx context.Context, result db.TaskResult) error {
	s.logger.Info("Yangi natija yaratish", "task_id", result.TaskID)

	if result.TaskID == "" {
		s.logger.Error("Task ID bo'sh bo'lishi mumkin emas")
		return errors.New("task ID majburiy")
	}

	if err := s.storage.TaskResult().CreateResult(ctx, result); err != nil {
		s.logger.Error("Natijani saqlashda xato", "error", err)
		return fmt.Errorf("natijani saqlashda xato")
	}
	return nil
}

// UpdateResult - Natijani yangilash (faqat file_url va git_url uchun)
func (s *ResultService) UpdateResult(ctx context.Context, resultID string, updates map[string]string) error {
	s.logger.Info("Natijani yangilash", "result_id", resultID)

	existingResult, err := s.storage.TaskResult().GetResult(ctx, resultID)
	if err != nil {
		return fmt.Errorf("natija topilmadi: %w", err)
	}

	if fileURL, ok := updates["file_url"]; ok {
		existingResult.FileURL = fileURL
	}
	if gitURL, ok := updates["git_url"]; ok {
		existingResult.GitURL = gitURL
	}

	if err := s.storage.TaskResult().UpdateResult(ctx, existingResult); err != nil {
		s.logger.Error("Yangilashda xato", "error", err)
		return fmt.Errorf("yangilashda xato")
	}
	return nil
}

// DeleteResult - Natijani o'chirish
func (s *ResultService) DeleteResult(ctx context.Context, resultID string) error {
	s.logger.Info("Natijani o'chirish", "result_id", resultID)

	if err := s.storage.TaskResult().DeleteResult(ctx, resultID); err != nil {
		s.logger.Error("O'chirishda xato", "error", err)
		return fmt.Errorf("o'chirishda xato")
	}
	return nil
}
