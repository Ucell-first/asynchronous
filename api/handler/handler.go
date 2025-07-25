package handler

import (
	"asynchronous/service"
	"log/slog"

	"github.com/casbin/casbin/v2"
)

type Handler struct {
	User   *service.UserService
	Task   *service.TaskService
	Result *service.ResultService
	Log    *slog.Logger
	Casbin *casbin.Enforcer
}

type ErrorResp struct {
	Error string `json:"error"`
}

type SuccessResp struct {
	Message string `json:"message"`
}
