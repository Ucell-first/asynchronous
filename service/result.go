package service

import (
	"asynchronous/storage"
	"asynchronous/storage/postgres"
	"database/sql"
	"log/slog"
)

type ResultService struct {
	Storage storage.IStorage
	Logger  *slog.Logger
}

func NewResultService(db *sql.DB, Logger *slog.Logger) *ResultService {
	return &ResultService{
		Storage: postgres.NewPostgresStorage(db),
		Logger:  Logger,
	}
}
