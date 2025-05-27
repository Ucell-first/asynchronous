package main

import (
	"asynchronous/api"
	"asynchronous/api/handler"
	"asynchronous/config"
	"asynchronous/logs"
	"asynchronous/service"
	"asynchronous/storage/postgres"
	"log"
	"log/slog"
)

func main() {
	cfg := config.Load()

	logger := logs.NewLogger()

	db, err := postgres.ConnectionDb()
	if err != nil {
		log.Fatal("Databasega ulanishda xato: ", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Database ulanishni yopishda xato", "error", err)
		}
	}()

	strg := postgres.NewPostgresStorage(db)

	defer strg.Close()

	userService := service.NewUserService(strg, logger)
	taskService := service.NewTaskService(strg, logger, cfg.Worker.WorkerCount)
	resultService := service.NewResultService(db, logger)

	hand := NewHandler(userService, taskService, resultService, logger)
	router := api.Router(hand)
	err = router.Run(cfg.Server.ROUTER)
	if err != nil {
		log.Fatal(err)
	}
}

func NewHandler(
	userService *service.UserService,
	taskService *service.TaskService,
	resultService *service.ResultService,
	logger *slog.Logger,
) *handler.Handler {
	return &handler.Handler{
		User:   userService,
		Task:   taskService,
		Result: resultService,
		Log:    logger,
	}
}
