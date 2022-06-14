package managers

import "github.com/paramonies/ya-gophermart/internal/server/dto"

type AppManagers interface {
	UserManager() UserManager
	OrderManager() OrderManager
	AccrualManager() AccrualManager
}

type UserManager interface {
	Create(req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	Login(req *dto.LoginUserRequest) (*dto.LoginUserResponse, error)
}

type OrderManager interface {
	Load(req *dto.LoadOrderRequest) (*dto.LoadOrderResponse, error)
	ListProcessed(req *dto.ListOrdersRequest) (*dto.ListOrdersResponse, error)
	GetBalance(req *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error)
}

type AccrualManager interface {
	PayOrder(req *dto.PayOrderRequest) (*dto.PayOrderResponse, error)
	GetOrders(req *dto.GetOrdersRequest) (*dto.GetOrdersResponse, error)
}
