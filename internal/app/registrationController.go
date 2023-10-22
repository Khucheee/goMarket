package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Khucheee/goMarket/internal/auth"
	"net/http"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// регистрация
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	//получаем прару пароль/логин.
	//каждый логин должен быть уникальным
	//после успешной регистрации должа происходить автоматическая аутентификация
	//для аутентификации выдавать куку или http:authorization
	//200 пользователь успешно зарегистрировался и прошел аутентификацию
	//400 неверный формат запроса
	//409 логин уже занят
	//500 внутренняя ошибка сервера
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var credentials Credentials    //создаем структуру в которую парсим полученный json
	var buf bytes.Buffer           //создаем буфер для получение тела запроса
	_, err := buf.ReadFrom(r.Body) //читаем тело запроса в буфер
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.Unmarshal(buf.Bytes(), &credentials) //парсим тело в нашу структуру
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if credentials.Login == "" || credentials.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	exist := c.storage.CheckLoginExist(credentials.Login) //проверяем уникальность логина
	if exist {
		w.WriteHeader(http.StatusConflict)
		return
	} //если логин существует то возвращаем ошибку
	userID, err := c.storage.CreateUser(credentials.Login, credentials.Password) //если логин уникальный - выдаем пользователю uuid
	if err != nil {
		fmt.Println("Что-то сломалось на этапе создания пользователя в базе", err)
	}
	//теперь нам нужно создать ему куку и выдать, чтобы он считался авторизованным
	authCookie, err := auth.MakeAuthCookie(userID)
	if err != nil {
		fmt.Println("Что-то упало на создании новой куки", err)
	}
	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
}

// авторизация
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	//получаем пару логин/пароль
	//200 - пользователь успешно аутентифицирован
	//400 - неверный формат запроса
	//500 - внутренняя ошибка сервера
	var credentials Credentials    //создаем структуру в которую парсим полученный json
	var buf bytes.Buffer           //создаем буфер для получение тела запроса
	_, err := buf.ReadFrom(r.Body) //читаем тело запроса в буфер
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.Unmarshal(buf.Bytes(), &credentials) //парсим тело в нашу структуру
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := c.storage.Login(credentials.Login, credentials.Password)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	authCookie, err := auth.MakeAuthCookie(userID)
	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
	//может добавить u allredy login?
}
