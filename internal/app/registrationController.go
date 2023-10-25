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
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	var credentials Credentials    //создаем структуру в которую парсим полученный json
	var buf bytes.Buffer           //создаем буфер для получение тела запроса
	_, err := buf.ReadFrom(r.Body) //читаем тело запроса в буфер
	if err != nil {
		fmt.Println("ошибка при чтении тела в ручке регистрации", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	err = json.Unmarshal(buf.Bytes(), &credentials) //парсим тело в нашу структуру
	if err != nil {
		fmt.Println("ошибка при парсинге тела в ручке регистрации", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	if credentials.Login == "" || credentials.Password == "" {
		fmt.Println("Логин или пароль пустые, неверный формат запроса")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	exist := c.storage.CheckLoginExist(credentials.Login) //проверяем уникальность логина
	if exist {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Логин уже занят, выбирете другой"))
		return
	} //если логин существует то возвращаем ошибку
	userID, err := c.storage.CreateUser(credentials.Login, credentials.Password) //если логин уникальный - выдаем пользователю uuid
	if err != nil {
		fmt.Println("Что-то сломалось на этапе создания пользователя в базе", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//теперь нам нужно создать ему куку и выдать, чтобы он считался авторизованным
	authCookie, err := auth.MakeAuthCookie(userID)
	if err != nil {
		fmt.Println("Что-то упало на создании новой куки", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		fmt.Println("ошибка при чтении тела в ручке авторизации", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	err = json.Unmarshal(buf.Bytes(), &credentials) //парсим тело в нашу структуру
	if err != nil {
		fmt.Println("ошибка при парсинге тела в ручке авторизации", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат запроса"))
		return
	}
	userID := c.storage.Login(credentials.Login, credentials.Password)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Неверная пара логин/пароль"))
		return
	}
	authCookie, err := auth.MakeAuthCookie(userID)
	if err != nil {
		fmt.Println("Произошла ошибка при создании куки для авторизации")
	}
	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
	//может добавить u allredy login?
}
