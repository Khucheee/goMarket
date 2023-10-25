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
	config := config.NewConfig()          //устанавливаем конфиг
	storage := storage.NewStorage(config) //создаем хранилище

	accrualWorker, accrualChannel := concurrency.NewAccrualWorker(storage, config)
	ordersWorker := concurrency.NewOrdersWorker(storage, 5, accrualChannel)
	accrualWorker.Start(context.Background())
	ordersWorker.Start(context.Background())

	logger := logger.NewLogger()
	logger.CreateSuggarLogger()
	controller := app.NewController(storage, config, logger) //создаем оператор //переименовать в контроллер обратно))
	router := chi.NewRouter()                                //создаем роутер
	router.Mount("/", controller.Route())
	err := http.ListenAndServe(config.RunAddress, router) //запускаем сервер
	if err != nil {
		fmt.Println("Ошибка при запуске сервера", err)
	}
}

//сначала меняю логику с сохранением и учетом заявок
//заодно чищу код
//затем меняю воркеры
//затем в выходные прорабатываю ошибки
