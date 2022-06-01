package store

import (
	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

type Connector interface {
	Users() UserRepoIf
	Accruals() AccrualRepoIf
	Withdrawns() WithdrawnRepoIf
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
	GetOrderByOrderNumber(orderNumber int) (*dto.Order, error)
	GetOrderByUserID(id string) (*[]dto.Order, error)
	UpdateOrder(order dto.ProviderOrder) error
	SelectOrders(userID string) ([]dto.Order, error)
}

type WithdrawnRepoIf interface {
	GetUserWithdrawnSum(userID int) (*float64, error)
}
