package store

import (
	"github.com/paramonies/ya-gophermart/internal/dto"
)

type Repository interface {
	CreateUser(userName, hash string) error
	GetHashedPassword(login string) (*string, error)

	CreateOrder(orderNumber int, id string) error
	UpdateOrder(order dto.ProviderOrder) error
	SelectOrders(userID string) ([]dto.Order, error)
}
