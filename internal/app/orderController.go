package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"github.com/Khucheee/goMarket/internal/concurrency"
	"log"
	"net/http"
)

// загрузка номера заказа
func (c *Controller) EvaluateOrder(w http.ResponseWriter, r *http.Request) {
	//номер заказа - цифры произвольной длины
	//Номер заказа можно проверить на корректность с помощью алгоритма Луна
	//200 - номер заказа уже был загружен этим пользователям
	//202 - новый номер заказа был принят в обработку
	//400 - неверный формат запроса
	//401 - пользователь не аутентифицирован
	//409 - номер заказа был загружен другим пользователем
	//422 - неверный формат номера заказа
	//500 - внутренняя ошибка сервера
	/*
			Тут мне нужно получить ответ от системы рассчета баллов,
			//должен запрашивать айдишники всех заказов в промежуточном статусе каждые n секунд, а потом отправлять их через
		канал в воркер, чтобы все это обрабатывать
	*/
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var buf bytes.Buffer //создаем буфер для получение тела запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		fmt.Println("Что-то упало на чтении тела запроса при расчете заказа", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := buf.String()
	if len(orderID) != 10 {
		fmt.Println("Неправильный формат номера заказа")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	owner := c.storage.CheckOrderOwner(orderID)
	//если создатель не вы
	if owner != userID && owner != "" {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if owner == userID {
		w.WriteHeader(http.StatusOK)
		return
	}
	//дальше, если такого заказа еще не было, то идем к сервису рассчетов, получаем из него данные и записываем в базу
	fmt.Println("передаем значение в воркеры")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Заказ притят в обработку"))
	orderForWorker := concurrency.OrderForWorker{OrderID: orderID, UserID: userID}
	*c.accrualChannel <- orderForWorker

	//c.CalculateOrder(orderID, userID)
}

// получение списка загруженных номеров заказов
func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {

	//хендлер доступен только авторизованному пользователю
	//номера заказа в выдаче должны быть отсортированы по времени загрузки от самых старых к самым новым
	//формат даты RFC3339
	//статусы
	//new - загружен в систему, но не попал в обработку
	//processing - вознаграждение за заказ рассчитывается
	//invalid - система рассчета вознаграждений отказала в расчёте
	//processed - данные по заказу проверены и информация о расчёте успешно получена
	//200 - успешная обработка ответа
	/*
		[
			{
				"number":"",
				"status":"",
				"accural":1,
				"uploaded_at":"date"
			},...
		]
	*/
	//204 - нет данных для ответа
	//401 - пользователь не авторизован
	//500 - внутренняя ошибка сервера
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userOrdersInfo := c.storage.GetUserOrders(userID)
	resp, err := json.Marshal(userOrdersInfo) //тут собираем их в jsonkу
	if err != nil {
		log.Printf("AllUserLinks: could not encode json \n %#v \n %#v \n\n", err, userOrdersInfo)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
