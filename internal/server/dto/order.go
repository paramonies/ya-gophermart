package dto

import "github.com/paramonies/ya-gophermart/internal/store/dto"

type LoadOrderRequest struct {
	User        *dto.User
	OrderNumber int
}

type LoadOrderResponse struct {
	StatusCode int
	MsgUser    string
	MsgLog     string
}

type ListOrdersRequest struct {
	UserID string
}

type ListOrdersResponse struct {
	Orders     *[]dto.OrderAccrual
	StatusCode int
	MsgUser    string
	MsgLog     string
}

type GetBalanceRequest struct {
	UserID string
}

type GetBalanceResponse struct {
	CurrentSum   *float64
	WithdrawnSum *float64
	Orders       *[]dto.OrderAccrual
	StatusCode   int
	MsgUser      string
	MsgLog       string
}
