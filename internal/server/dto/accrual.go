package dto

import "github.com/paramonies/ya-gophermart/internal/store/dto"

type PayOrderRequest struct {
	UserID      string
	OrderNumber int
	Price       float64
}

type PayOrderResponse struct {
	StatusCode int
	MsgUser    string
	MsgLog     string
}

type GetOrdersRequest struct {
	UserID string
}

type GetOrdersResponse struct {
	Orders     *[]dto.Order
	StatusCode int
	MsgUser    string
	MsgLog     string
}
