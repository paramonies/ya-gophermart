package store

import (
	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

type Connector interface {
	Users() UserRepoIf
	Accruals() AccrualRepoIf
	Orders() OrderRepoIf
}

type UserRepoIf interface {
	GetByName(name string) (*dto.User, error)
	Create(userName, hash string) error
	SetToken(userName, token string) error
	GetByToken(token string) (*dto.User, error)
	GetHashedPassword(login string) (*string, error)
}

type AccrualRepoIf interface {
	LoadOrder(orderNumber int, userID string) error
	GetOrderByOrderNumber(orderNumber int) (*dto.OrderAccrual, error)
	GetOrderByUserID(id string) (*[]dto.OrderAccrual, error)
	GetPendingOrdersByUserID(id string) (*[]dto.OrderAccrual, error)
	GetPendingOrders() (*[]dto.OrderAccrual, error)
	UpdateAccrual(or dto.ProviderOrder) error
	SelectOrders(userID string) ([]dto.OrderAccrual, error)
	GetOrdersAccrualForUser(userID string) (*float64, error)
}

type OrderRepoIf interface {
	GetOrdersPriceForUser(userID string) (*float64, error)
	Register(userID string, orderNumber string, price float64) error
	GetOrdersByUserID(userID string) (*[]dto.Order, error)
}
