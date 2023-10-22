package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"math"
	"time"
)

type UserOrderInfo struct {
	OrderID    string  `json:"number"`
	Status     string  `json:"status"`
	Amount     float64 `json:"accural"`
	Created_at string  `json:"uploaded_at"`
}

type WithdrawalsHistory struct {
	OrderID    string  `json:"order"`
	Amount     float64 `json:"sum"`
	Created_at string  `json:"processed_at"`
}

// "host=localhost user=postgres password=ALFREd2002 dbname=postgres sslmode=disable"
const migrationFolder = "file://internal/storage/migrations/"

type Postgres struct {
	DatabaseURI  string
	dbConnection *sql.DB
}

func (p *Postgres) Initialize() {
	dbConnection, err := sql.Open("pgx", p.DatabaseURI)
	if err != nil {
		fmt.Println("Упала база на создании соединения", err)
	}
	//миграция
	driver, err := postgres.WithInstance(dbConnection, &postgres.Config{})
	if err != nil {
		fmt.Println("Упала база на получении драйвера для миграции", err)
	}
	m, err := migrate.NewWithDatabaseInstance(migrationFolder, "postgres", driver)
	if err != nil {
		fmt.Println("База упала на создании миграции:", err)
	}
	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			fmt.Println("Упала миграция при поднятии", err)
		}
	}
	//сохраняем подключение
	p.dbConnection = dbConnection
}

func (p *Postgres) CheckLoginExist(login string) bool {
	//запрос в базу
	var result string
	row := p.dbConnection.QueryRowContext(context.Background(),
		"SELECT user_id FROM usr WHERE login=$1", login)
	row.Scan(&result)
	//если вернулось пусто, то return false
	if result != "" {
		return true
	}
	return false
}

func (p *Postgres) CreateUser(login, password string) (string, error) {
	userID := uuid.New().String()
	_, err := p.dbConnection.ExecContext(context.Background(),
		"INSERT INTO usr VALUES($1,$2,$3)",
		userID, login, password)
	if err != nil {
		fmt.Println("Что-то упало при создании пользователя в базе:", err)
	}
	return userID, nil
}

func (p *Postgres) Login(login, password string) string {
	var result string
	row := p.dbConnection.QueryRowContext(context.Background(),
		"SELECT user_id FROM usr WHERE login=$1 and password=$2", login, password)
	row.Scan(&result)
	//если вернулось пусто, то return false
	return result
}

func (p *Postgres) CheckOrderOwner(orderID string) string {
	var result string
	row := p.dbConnection.QueryRowContext(context.Background(),
		"SELECT user_id FROM orders WHERE order_id=$1", orderID)
	row.Scan(&result)
	return result
}

func (p *Postgres) CreateOrder(orderID, userID, status string, amount float64) {
	_, err := p.dbConnection.ExecContext(context.Background(),
		"INSERT INTO orders VALUES($1,$2,$3,$4,now(),now())",
		orderID, userID, status, amount)
	if err != nil {
		fmt.Println("Что-то упало при создании заказа в базе:", err)
	}
}

func (p *Postgres) UpdateOrder(orderID, status string, amount float64) {
	query := "UPDATE orders SET status=$1,amount=$2,updated_at=now() WHERE order_id=$3"
	_, err := p.dbConnection.ExecContext(context.Background(), query, status, amount, orderID)
	if err != nil {
		log.Println("Что-то упало на обновлении заказа в базе", err)
	}
}

func (p *Postgres) GetUserOrders(userID string) []UserOrderInfo {
	orders := []UserOrderInfo{}
	rows, err := p.dbConnection.QueryContext(context.Background(),
		"SELECT order_id,status,amount,created_at FROM orders WHERE user_id = $1 order by created_at", userID)
	if err != nil {
		fmt.Println("Что-то упало на запросе заказов пользователя", err)
		return orders
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("Ошибка в чтении строк в таблице:", err)
	}
	var tmp UserOrderInfo
	for rows.Next() {
		err = rows.Scan(&tmp.OrderID, &tmp.Status, &tmp.Amount, &tmp.Created_at)
		if err != nil {
			fmt.Println("Что-то упало на сканировании полученных строк по заказам пользователя:", err)
		}
		t, err := time.Parse(time.RFC3339, tmp.Created_at)
		if err != nil {
			fmt.Println("Не получилось распарсить дату и время в нужный формат", err)
		}
		tme := t.Format(time.RFC3339)
		orders = append(orders, UserOrderInfo{tmp.OrderID, tmp.Status, tmp.Amount, tme})
	}
	return orders
}

func (p *Postgres) RegisterIncomeTransaction(userID, orderID string, amount float64) {
	operationType := "INCOME"
	_, err := p.dbConnection.ExecContext(context.Background(),
		"INSERT INTO account_transaction VALUES($1,$2,$3,$4,now())", userID, operationType, orderID, amount)
	if err != nil {
		fmt.Println("Что-то упало при создании поступления средств на кошелек в базе:", err)
	}
}

func (p *Postgres) GetWalletInfo(userID string) (float64, float64) {
	var incomes float64
	row := p.dbConnection.QueryRowContext(context.Background(),
		"SELECT sum(amount) FROM account_transaction WHERE user_id = $1 and operation_type='INCOME'", userID)
	row.Scan(&incomes)

	var outcomes float64
	row = p.dbConnection.QueryRowContext(context.Background(),
		"SELECT sum(amount) FROM account_transaction WHERE user_id = $1 and operation_type='OUTCOME'", userID)
	row.Scan(&outcomes)
	balance := incomes - outcomes
	//если вернулось пусто, то return false
	fmt.Println("balance:", balance, "incomes:", incomes, "outcomes:", outcomes)
	return math.Round(balance*100) / 100, outcomes
}

func (p *Postgres) RegisterOutcomeTransaction(userID, orderID string, withdrawal float64) bool {
	balance, _ := p.GetWalletInfo(userID)
	if withdrawal > balance {
		return false
	}
	operationType := "OUTCOME"
	_, err := p.dbConnection.ExecContext(context.Background(),
		"INSERT INTO account_transaction VALUES($1,$2,$3,$4,now())", userID, operationType, orderID, withdrawal)
	if err != nil {
		fmt.Println("Что-то упало при создании операции на списание средств c кошелька в базе:", err)
	}
	return true
}

func (p *Postgres) GetUserWithdrawals(userID string) []WithdrawalsHistory {
	withdrawals := []WithdrawalsHistory{}
	rows, err := p.dbConnection.QueryContext(context.Background(),
		"SELECT order_id,amount,created_at FROM account_transaction WHERE user_id = $1 order by created_at", userID)
	if err != nil {
		fmt.Println("Что-то упало на запросе заказов пользователя", err)
		return withdrawals
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("Ошибка в чтении строк в таблице:", err)
	}
	var tmp WithdrawalsHistory
	for rows.Next() {
		err = rows.Scan(&tmp.OrderID, &tmp.Amount, &tmp.Created_at)
		if err != nil {
			fmt.Println("Что-то упало на сканировании полученных строк по заказам пользователя:", err)
		}
		t, err := time.Parse(time.RFC3339, tmp.Created_at)
		if err != nil {
			fmt.Println("Не получилось распарсить дату и время в нужный формат", err)
		}
		tme := t.Format(time.RFC3339)
		withdrawals = append(withdrawals, WithdrawalsHistory{tmp.OrderID, tmp.Amount, tme})
	}
	return withdrawals
}

//как будет выглядеть бд:
//номер заказа - строка - начиленные бонусы - числа
//account_transaction
// user id | operation_type | status | amount | date

//user
//user id|login|password

//таблица с заказами orders
// order id |status| description(состав заказа или user id)
