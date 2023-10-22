package storage

import (
	"github.com/Khucheee/goMarket/internal/config"
)

type Storage interface {
	Initialize()
	CheckLoginExist(string) bool
	CreateUser(string, string) (string, error)
	Login(string, string) string
	CheckOrderOwner(string) string
	CreateOrder(string, string, string, float64) bool
	UpdateOrder(string, string, float64)
	GetUserOrders(string) []UserOrderInfo
	RegisterIncomeTransaction(string, string, float64)
	GetWalletInfo(string) (float64, float64)
	RegisterOutcomeTransaction(string, string, float64) bool
	GetUserWithdrawals(string) []WithdrawalsHistory
}

func NewStorage(config *config.Config) Storage {
	storage := &Postgres{DatabaseURI: config.DatabaseURI}
	storage.Initialize()
	return storage
}
