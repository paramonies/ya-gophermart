package store

import "github.com/paramonies/ya-gophermart/internal/dto"

type Connector interface {
	Users() UserRepoIf
	Orders() OrderRepoIf
	Withdrawns() WithdrawnRepoIf
}

type UserRepoIf interface {
	CreateUser(userName, hash string) error
	GetHashedPassword(login string) (*string, error)
}

type OrderRepoIf interface {
	CreateOrder(orderNumber int, id string) error
	UpdateOrder(order dto.ProviderOrder) error
	SelectOrders(userID string) ([]dto.Order, error)
}

type WithdrawnRepoIf interface {
	GetUserWithdrawnSum(userID int) (*float64, error)
}
