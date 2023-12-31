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
	"time"
)

type worker struct {
	wg             *sync.WaitGroup
	cancelFunc     context.CancelFunc
	storage        storage.Storage
	accrualChannel chan OrderForWorker
	config         *config.Config
}

type Worker interface {
	Start(pctx context.Context)
	Stop()
}

type OrderForWorker struct {
	OrderID string
	UserID  string
}

type AccrualServiceResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewWorker(storage storage.Storage, config *config.Config) (Worker, chan OrderForWorker) {
	w := worker{
		wg:             new(sync.WaitGroup),
		storage:        storage,
		accrualChannel: make(chan OrderForWorker),
		config:         config,
	}
	return &w, w.accrualChannel
}

func (w *worker) Start(pctx context.Context) {
	ctx, cancelFunc := context.WithCancel(pctx)
	w.cancelFunc = cancelFunc

	for i := 0; i <= 30; i++ {
		w.wg.Add(1)
		go w.spawnWorkers(ctx)
	}

}

func (w *worker) Stop() {
	w.cancelFunc()
	w.wg.Wait()
}

func (w *worker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case value := <-w.accrualChannel:
			fmt.Println("Произошло чтение из канала, получены значения:", value)
			w.CalculateOrder(value.OrderID, value.UserID)

		}
	}
}

func (w *worker) CalculateOrder(orderID, userID string) {
	fmt.Println("выполняется запрос к рассчетному сервису")
	//response, err := http.Get("http://" + w.config.AccuralSystemAddress + "/api/orders/" + orderID)
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, w.config.AccuralSystemAddress+"/api/orders/"+orderID, nil)
	if err != nil {
		fmt.Println("Что-то упало на запросе к системе рассчета", err)
	}
	fmt.Println("Вывожу адресс по которому мы стучимся к системе рассчета:", w.config.AccuralSystemAddress+"/api/orders/"+orderID)
	request.Header.Add("Content-Length", "0")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Упал запрос к accural:", err)
	}
	fmt.Println("Статус код ответа сервиса рассчета:", response.StatusCode)
	orderData := GetResponseBody(response)
	fmt.Println("Создаем заказ в базе")
	ok := w.storage.CreateOrder(orderID, userID, orderData.Status, orderData.Accrual)
	//пока колхозная проверка на существование заказа в базе
	if !ok {
		return
	}
	if orderData.Status == "PROCESSED" {
		fmt.Println("Происходит регистрация транзакции, начисление денег на кошелек")
		w.storage.RegisterIncomeTransaction(userID, orderID, orderData.Accrual)
	}
	//если статус заказа промежуточный
	if orderData.Status == "REGISTERED" || orderData.Status == "PROCESSING" || orderData.Status == "" {
		//каждые 1 секунд запрашиваем новые данные по этому заказу
		for {
			fmt.Println("Запущена джоба по обновлению данных заказа")
			time.Sleep(time.Second)
			//response, err := http.Get("http://" + w.config.AccuralSystemAddress + "/api/orders/" + orderID)
			request, err = http.NewRequest(http.MethodGet, w.config.AccuralSystemAddress+"/api/orders/"+orderID, nil)
			if err != nil {
				fmt.Println("Что-то упало на запросе к системе рассчета", err)
			}
			request.Header.Add("Content-Length", "0")
			response, err = client.Do(request)
			if err != nil {
				fmt.Println("Что-то упало на запросе к системе рассчета", err)
			}
			orderData = GetResponseBody(response)
			//если получаем конечный статус, то обновляем данные по заказу и прерывем цикл
			if response.Status == "PROCESSED" {
				w.storage.UpdateOrder(orderID, orderData.Status, orderData.Accrual)
				fmt.Println("Происходит регистрация транзакции, начисление денег на кошелек")
				w.storage.RegisterIncomeTransaction(userID, orderID, orderData.Accrual)
				break
			}
			if response.Status == "INVALID" {
				w.storage.UpdateOrder(orderID, orderData.Status, orderData.Accrual)
				break
			}
		}
		//после того как обновили данные, выходим из хендлера
		return
	}
	//если же статус заказа оказался конечным, то записываем полученные значения в базу

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
