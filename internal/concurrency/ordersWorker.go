package concurrency

import (
	"context"
	"github.com/Khucheee/goMarket/internal/storage"
	"runtime"
	"sync"
	"time"
)

type OrderWorker struct {
	wg             *sync.WaitGroup
	cancelFunc     context.CancelFunc
	storage        storage.Storage
	accrualChannel chan string
	forTimer       int
}

func NewOrdersWorker(storage storage.Storage, seconds int, accrualChannel chan string) Worker {
	w := OrderWorker{
		wg:             new(sync.WaitGroup),
		storage:        storage,
		accrualChannel: accrualChannel,
		forTimer:       seconds,
	}
	return &w
}

func (w *OrderWorker) Start(pctx context.Context) {
	ctx, cancelFunc := context.WithCancel(pctx)
	w.cancelFunc = cancelFunc
	for i := 0; i < runtime.NumCPU(); i++ {
		w.wg.Add(1)
		go w.spawnWorkers(ctx)
	}

}

func (w *OrderWorker) Stop() {
	w.cancelFunc()
	w.wg.Wait()
}

func (w *OrderWorker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	duration := time.Duration(w.forTimer)
	t := time.NewTicker(time.Second * duration)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.doWork()
		}
	}
}

func (w *OrderWorker) doWork() {
	orders := w.storage.GetOrdersForUpdate()
	for _, order := range orders {
		w.accrualChannel <- order
	}
}
