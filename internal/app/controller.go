package app

import (
	"github.com/Khucheee/goMarket/internal/concurrency"
	"github.com/Khucheee/goMarket/internal/config"
	"github.com/Khucheee/goMarket/internal/logger"
	"github.com/Khucheee/goMarket/internal/storage"
	"net/http"
	"time"
)

type Controller struct {
	storage        storage.Storage
	config         *config.Config
	worker         concurrency.Worker
	accrualChannel *chan concurrency.OrderForWorker
	logger         *logger.Logger
	OrdersCache    map[string]string
}

func NewController(storage storage.Storage, config *config.Config, worker concurrency.Worker,
	accrualChannel chan concurrency.OrderForWorker, logger *logger.Logger) *Controller {
	controller := Controller{
		storage:        storage,
		config:         config,
		worker:         worker,
		accrualChannel: &accrualChannel,
		logger:         logger,
		OrdersCache:    make(map[string]string),
	}
	return &controller
}

func (b *Controller) WithLogging(h http.Handler) http.Handler {
	logfn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &logger.ResponseData{Status: 0, Size: 0}
		lw := logger.LoggingResponseWriter{w, responseData}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		b.logger.Sugar.Infoln(
			"URI", r.RequestURI,
			"duration", duration,
			"method", r.Method,
			"status", responseData.Status,
			"size", responseData.Size)
	}
	return http.HandlerFunc(logfn)
}
