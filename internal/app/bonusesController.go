package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"log"
	"net/http"
)

type UserBonusesInfo struct {
	Balance  float64 `json:"current"`
	Outcomes float64 `json:"withdrawn"`
}

type UseUserBonuses struct {
	OrderID    string  `json:"order"`
	Withdrawal float64 `json:"sum"`
}

// получение текущего баланса пользователя
func (c *Controller) SeeUserBonuses(w http.ResponseWriter, r *http.Request) {
	//доступен только авторизованному пользователю
	//должны содержаться данные о текущей сумме баллов лояльности
	//должны содержаться данные о сумме использованных за весь период регистрации баллов
	/*
		{
			"current":float,
			"withdrawn":float,
		}
	*/
	//
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userBonusesInfo := UserBonusesInfo{}
	userBonusesInfo.Balance, userBonusesInfo.Outcomes = c.storage.GetWalletInfo(userID)
	resp, err := json.Marshal(userBonusesInfo) //тут собираем их в jsonkу
	if err != nil {
		log.Printf("Ошибка при сборке json с данными баланса кошелька", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// запрос на списание средств
func (c *Controller) UseUserBonuses(w http.ResponseWriter, r *http.Request) {
	//доступен только для авторизованных пользователей
	//номер заказа - гипотетический номер заказа пользователя в счет которого спписываются пользователи
	//для успешного списания достаточно успешной реализации запроса
	/*
		{
			"order":"",
			"sum":751
		}
	*/
	//200 - успешная обработка запроса
	//401 - пользователь не авторизован
	//402 - на счету недостаточно средств
	//422 - неверный номер заказа
	//внутренняя ошибка сервера
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var useUserBonuses UseUserBonuses //создаем структуру в которую парсим полученный json
	var buf bytes.Buffer              //создаем буфер для получение тела запроса
	_, err = buf.ReadFrom(r.Body)     //читаем тело запроса в буфер
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.Unmarshal(buf.Bytes(), &useUserBonuses) //парсим тело в нашу структуру
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(useUserBonuses.OrderID) != 10 {
		fmt.Println("Неправильный формат номера заказа")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	ok := c.storage.RegisterOutcomeTransaction(userID, useUserBonuses.OrderID, useUserBonuses.Withdrawal)
	if !ok {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte("У вас недостаточно бонусов для списания"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Бонусы успешно списаны"))

}

// получение информации о выводе средств
func (c *Controller) SeeBonusesAccountHistory(w http.ResponseWriter, r *http.Request) {
	//доступен только авторизованным пользователем
	//факты выводов в выдаче должны быть отсортированы по времени вывода от самых старых к самым новым
	//формат даты - RFC3339
	//200 - успешная обработка запроса
	/*[
		{
			"order":"238",
			"sum":500,
			"processed_at":"date"
		}
	  ]*/
	//204 - нет ни одного списания
	//401 - пользователь не авторизован
	//500 - внутренняя ошибка сервера
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userWithdrawalsInfo := c.storage.GetUserWithdrawals(userID)
	resp, err := json.Marshal(userWithdrawalsInfo)
	if err != nil {
		log.Printf("AllUserLinks: could not encode json \n %#v \n %#v \n\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
