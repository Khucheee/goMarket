package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"github.com/Khucheee/goMarket/internal/luhn"
	"net/http"
	"strconv"
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

	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("пользователь не аутентифицирован"))
		return
	}
	userBonusesInfo := UserBonusesInfo{}
	userBonusesInfo.Balance, userBonusesInfo.Outcomes = c.storage.GetWalletInfo(userID)
	resp, err := json.Marshal(userBonusesInfo) //тут собираем их в jsonkу
	if err != nil {
		fmt.Println("Ошибка при сборке json с данными баланса кошелька", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// запрос на списание средств
func (c *Controller) UseUserBonuses(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.ParseUserFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("пользователь не аутентифицирован"))
		return
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Сообщение должно быть отправлено в формате json"))
		return
	}
	var useUserBonuses UseUserBonuses
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(buf.Bytes(), &useUserBonuses) //парсим тело в нашу структуру
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	check, err := strconv.Atoi(useUserBonuses.OrderID)
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
		w.Write([]byte("Пользователь не аутентифицирован"))
		return
	}
	userWithdrawalsInfo := c.storage.GetUserWithdrawals(userID)
	if len(userWithdrawalsInfo) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Вы еще ни разу не использовали бонусы"))
		return
	}
	resp, err := json.Marshal(userWithdrawalsInfo)
	if err != nil {
		fmt.Println("Не получилось собрать json для возврата истории списаний", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
