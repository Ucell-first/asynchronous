package service

import (
	models "asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"fmt"
	"log/slog"
)

type UserService struct {
	Storage storage.IStorage
	Logger  *slog.Logger
}

func NewUserService(db storage.IStorage, Logger *slog.Logger) *UserService {
	return &UserService{
		Storage: db,
		Logger:  Logger,
	}
}

func (s *UserService) Register(ctx context.Context, req models.User) (string, error) {
	s.Logger.Info("Register rpc methos is working")
	resp, err := s.Storage.User().CreateUser(ctx, req)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("registration error: %v", err))
		return "", err
	}
	s.Logger.Info("Register rpc method finished")
	return resp, nil
}
