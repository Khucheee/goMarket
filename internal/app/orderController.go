package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"github.com/Khucheee/goMarket/internal/concurrency"
	"github.com/Khucheee/goMarket/internal/luhn"
	"log"
	"net/http"
	"strconv"
)

// загрузка номера заказа
func (c *Controller) EvaluateOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("пользователь не аутентифицирован"))
		return
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
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
	check, err := strconv.Atoi(orderID)
	if err != nil {
		fmt.Println("Не получилось перевести строку в инт для проверки номера заказа", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("Номер заказа должен быть целым числом"))
		return
	}
	if !luhn.Valid(check) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("Неверный формат номера заказа"))
		return
	}
	owner := c.storage.CheckOrderOwner(orderID)
	//если создатель не вы
	if owner != userID && owner != "" { //если пользователь не наш и не пустой значит номер был загружен другим пользователем
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Номер заказа уже был загружен другим пользователем"))
		return
	}
	fmt.Println("сейчас для проверки сравним id из куки и id пользователя из базы, если совпадут, то 200")
	fmt.Println("userID:", userID, "owner:", owner)
	if owner == userID { //если пользователь наш, то мы уже загружали данный заказ
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Вы уже загружали данный заказ"))
		return
	}
	//дальше, если такого заказа еще не было, то идем к сервису рассчетов, получаем из него данные и записываем в базу
	fmt.Println("передаем значение в воркеры")

	orderForWorker := concurrency.OrderForWorker{OrderID: orderID, UserID: userID}
	*c.accrualChannel <- orderForWorker
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Новый номер заказа принят в обработку"))
	//c.CalculateOrder(orderID, userID)
}

// получение списка загруженных номеров заказов
func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("пользователь не аутентифицирован"))
		return
	}
	userOrdersInfo := c.storage.GetUserOrders(userID)
	if len(userOrdersInfo) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Вы еще не загрузили ни одного заказа"))
		return
	}
	fmt.Println("Тут находится информация по созданным пользователем заказам")
	fmt.Println(userOrdersInfo)
	resp, err := json.Marshal(userOrdersInfo) //тут собираем их в jsonkу
	if err != nil {
		log.Printf("AllUserLinks: could not encode json \n %#v \n %#v \n\n", err, userOrdersInfo)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
