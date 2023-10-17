package main

import (
	"github.com/go-chi/chi"
	"goMarket/internal/app"
	"net/http"
)

func main() {
	operator := app.NewOperator()
	router := chi.NewRouter()
	router.Mount("/", operator.Route())
	http.ListenAndServe(":8080", router)
}
