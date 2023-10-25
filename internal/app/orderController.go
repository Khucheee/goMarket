package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"github.com/Khucheee/goMarket/internal/luhn"
	"log"
	"net/http"
	"strconv"
)

// загрузка номера заказа
func (c *Controller) EvaluateOrder(w http.ResponseWriter, r *http.Request) {
	//проверяем авторизацию
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("пользователь не аутентифицирован"))
		return
	}
	//проверяем формат запроса
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	//читаем тело
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		fmt.Println("Что-то упало на чтении тела запроса при расчете заказа", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := buf.String()
	//проверяем формат номера заказа
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
	//лекгий лайфках с кэшэм
	//if c.OrdersCache[orderID] == userID {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("Вы уже загружали данный заказ"))
	//	return
	//}
	//c.OrdersCache[orderID] = userID
	//проверка на дубли
	owner := c.storage.CheckOrderOwner(orderID)
	//проверка на дубли чужих номеров
	if owner != userID && owner != "" { //если пользователь не наш и не пустой значит номер был загружен другим пользователем
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Номер заказа уже был загружен другим пользователем"))
		return
	}
	fmt.Println("сейчас для проверки сравним id из куки и id пользователя из базы, если совпадут, то 200")
	fmt.Println("userID:", userID, "owner:", owner)
	//проверка на дубли от самого себя
	if owner == userID {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Вы уже загружали данный заказ"))
		return
	}
	//как было раньше
	//orderForWorker := concurrency.OrderForWorker{OrderID: orderID, UserID: userID}
	//*c.accrualChannel <- orderForWorker

	//теперь сразу записывается значение в базу, записываю вместе с amount, так как требуется
	//возвращать число бонусов в ручке get api/user/orders
	//логика такая: воркер из файла concurrency/orderWorker каждое n количество секунд ходит в базу, чтобы получить
	//заказы в статусе NEW и PROCESSING, далее с помощью цикла закидывает значения в канал воркера, который живет в файле
	//accrualWorkers, тот в свою очередь читает и ходит в accrual
	c.storage.CreateOrder(orderID, userID, "NEW", 0)
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
