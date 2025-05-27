package handler

import (
	"asynchronous/service"
	"log/slog"
)

type Handler struct {
	User   *service.UserService
	Task   *service.TaskService
	Result *service.ResultService
	Log    *slog.Logger
}
