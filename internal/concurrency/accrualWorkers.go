package concurrency

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/config"
	"github.com/Khucheee/goMarket/internal/storage"
	"net/http"
	"sync"
)

type AccrualWorker struct {
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
	storage    storage.Storage
	ownChannel chan string
	config     *config.Config
}

type Worker interface {
	Start(pctx context.Context)
	Stop()
}

//type OrderForWorker struct {
//	OrderID string
//	UserID  string
//}

type AccrualServiceResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewAccrualWorker(storage storage.Storage, config *config.Config) (Worker, chan string) {
	w := AccrualWorker{
		wg:         new(sync.WaitGroup),
		storage:    storage,
		ownChannel: make(chan string),
		config:     config,
	}
	return &w, w.ownChannel
}

func (w *AccrualWorker) Start(pctx context.Context) {
	ctx, cancelFunc := context.WithCancel(pctx)
	w.cancelFunc = cancelFunc
	w.wg.Add(1)
	go w.spawnWorkers(ctx)

}

func (w *AccrualWorker) Stop() {
	w.cancelFunc()
	w.wg.Wait()
}

func (w *AccrualWorker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case value := <-w.ownChannel:
			fmt.Println("Произошло чтение из канала, получены значения:", value)
			w.CalculateOrder(value)

		}
	}
}

func (w *AccrualWorker) CalculateOrder(orderID string) {
	fmt.Println("выполняется запрос к рассчетному сервису")
	//response, err := http.Get("http://" + w.config.AccuralSystemAddress + "/api/orders/" + orderID)
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, w.config.AccuralSystemAddress+"/api/orders/"+orderID, nil)
	if err != nil {
		fmt.Println("Что-то упало на запросе к системе рассчета", err)
	}
	fmt.Println("Вывожу адресс по которому мы стучимся к системе рассчета:", w.config.AccuralSystemAddress+"/api/orders/"+orderID)
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Упал запрос к accural:", err)
	}
	fmt.Println("Статус код ответа сервиса рассчета:", response.StatusCode)
	orderData := GetResponseBody(response)
	//если статус processed, то обновляем статус в базе и создаем транзакцию
	if orderData.Status == "PROCESSED" {
		w.storage.UpdateOrder(orderID, orderData.Status, orderData.Accrual)
		fmt.Println("Происходит регистрация транзакции, начисление денег на кошелек")
		w.storage.RegisterIncomeTransaction(orderID, orderData.Accrual)
		return
	}
	if orderData.Status == "INVALID" {
		w.storage.UpdateOrder(orderID, orderData.Status, orderData.Accrual)
		return
	}
	//если статус заказа промежуточный
	if orderData.Status == "REGISTERED" || orderData.Status == "PROCESSING" {
		w.storage.UpdateOrder(orderID, "PROCESSING", orderData.Accrual)
		//после того как обновили данные, выходим из хендлера
		return
	}
	if orderData.Status == "" {
		return
	}
}

func GetResponseBody(response *http.Response) *AccrualServiceResponse {
	var orderData AccrualServiceResponse //создаем структуру в которую парсим полученный json
	var buf bytes.Buffer                 //создаем буфер для получение тела запроса
	fmt.Println("Выполняем чтение тела из ответа accrual")
	_, err := buf.ReadFrom(response.Body) //читаем тело запроса в буфер
	if err != nil {
		fmt.Println("Не получилсь прочитать тело ответа сервиса рассчета бонусов", err)
		return nil
	}
	err = json.Unmarshal(buf.Bytes(), &orderData) //парсим тело в нашу структуру
	if err != nil {
		fmt.Println("Не получилось распарсить json из тела ответа системы рассчета бонусов", err)
		return &orderData
	}
	return &orderData
}
