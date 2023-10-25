package main

import (
	"context"
	"fmt"
	"github.com/Khucheee/goMarket/internal/app"
	"github.com/Khucheee/goMarket/internal/concurrency"
	"github.com/Khucheee/goMarket/internal/config"
	"github.com/Khucheee/goMarket/internal/logger"
	"github.com/Khucheee/goMarket/internal/storage"
	"github.com/go-chi/chi"
	"net/http"
)

func main() {
	config := config.NewConfig()
	storage := storage.NewStorage(config)
	accrualWorker, accrualChannel := concurrency.NewAccrualWorker(storage, config)
	ordersWorker := concurrency.NewOrdersWorker(storage, 5, accrualChannel)
	accrualWorker.Start(context.Background())
	ordersWorker.Start(context.Background())
	logger := logger.NewLogger()
	logger.CreateSuggarLogger()
	controller := app.NewController(storage, config, logger)
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	err := http.ListenAndServe(config.RunAddress, router)
	if err != nil {
		fmt.Println("Ошибка при запуске сервера", err)
	}
}
