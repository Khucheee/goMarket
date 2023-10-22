package app

import (
	"github.com/Khucheee/goMarket/internal/compress"
	"github.com/go-chi/chi"
)

func (c *Controller) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Use(c.WithLogging)
	router.Use(compress.Gzip)
	router.Route("/", func(router chi.Router) {
		router.Route("/api", func(router chi.Router) {
			router.Route("/user", func(router chi.Router) {
				router.Post("/register", c.Register)
				router.Post("/login", c.Login)
				router.Route("/orders", func(router chi.Router) {
					router.Get("/", c.GetOrders)
					router.Post("/", c.EvaluateOrder)
				})
				router.Route("/balance", func(router chi.Router) {
					router.Get("/", c.SeeUserBonuses)
					router.Post("/withdraw", c.UseUserBonuses)
				})
				router.Get("/withdrawals", c.SeeBonusesAccountHistory)

			})
		})
	})
	return router
}
