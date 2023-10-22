package main

import (
	"context"
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
	worker, accrualChannel := concurrency.NewWorker(storage, config)
	worker.Start(context.Background())
	logger := logger.NewLogger()
	logger.CreateSuggarLogger()
	controller := app.NewController(storage, config, worker, accrualChannel, logger) //создаем оператор //переименовать в контроллер обратно))
	router := chi.NewRouter()                                                        //создаем роутер
	router.Mount("/", controller.Route())
	err := http.ListenAndServe(config.RunAddress, router) //запускаем сервер
	if err != nil {
		//тут будем логировать
	}
}

//норм ли будет, если буду исопльзовать только storage
//как происходит взаимодействие с системаой рассчета бонусов
//спросить про архитектуру (как воркеры с хранилищем должны работать)
//как по человечески сделать миграции!
//как учитывать промежуточные статусы в базе? (вот это большой вопрос)
//сделать джобу, которая каждые 5-10 секунд запрашивает данные по этому заказу и пока статус не будет конечным
//будет выполняться?
//тупые вопросы
//нужно ли хранить логин пароль в базе

//по бд буду использовать pgx
//миграции через goose, потому что golang migrate заставил меня поплакать
