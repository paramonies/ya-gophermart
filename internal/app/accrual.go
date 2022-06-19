package app

import (
	"errors"
	"fmt"
	"net/http"

	pgxv4 "github.com/jackc/pgx/v4"

	"github.com/paramonies/ya-gophermart/internal/managers"
	"github.com/paramonies/ya-gophermart/internal/server/dto"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/store/pgx"
	"github.com/paramonies/ya-gophermart/internal/utils"
)

type DefAccrualManager struct {
	storage store.Connector
}

func NewAccrualManager(storage store.Connector) managers.AccrualManager {
	return &DefAccrualManager{
		storage: storage,
	}
}

func (m *DefAccrualManager) PayOrder(req *dto.PayOrderRequest) (*dto.PayOrderResponse, error) {
	if !utils.Valid(req.OrderNumber) {
		msg := "order number is invalid according to Luhn"
		return &dto.PayOrderResponse{
			StatusCode: http.StatusUnprocessableEntity,
			MsgUser:    msg,
			MsgLog:     msg,
		}, errors.New(msg)
	}

	totalAccrual, err := m.storage.Accruals().GetOrdersAccrualForUser(req.UserID)
	if err != nil {
		return &dto.PayOrderResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get total accruals for all orders for user",
		}, err
	}

	totalPrice, err := m.storage.Orders().GetOrdersPriceForUser(req.UserID)
	if err != nil {
		return &dto.PayOrderResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get total prices for all orders for user",
		}, err
	}

	balance := *totalAccrual - *totalPrice
	if balance < req.Price {
		msg := "negative balance for user"
		return &dto.PayOrderResponse{
			StatusCode: http.StatusPaymentRequired,
			MsgUser:    "not enought money on your account",
			MsgLog:     msg,
		}, errors.New(msg)
	}

	err = m.storage.Orders().Register(req.UserID, fmt.Sprintf("%d", req.OrderNumber), float64(req.Price))
	if err != nil {
		if errors.Is(err, pgx.ErrConstraintViolationOrder) {
			msg := "order already registered"
			return &dto.PayOrderResponse{
				StatusCode: http.StatusOK,
				MsgUser:    msg,
				MsgLog:     msg,
			}, err
		}
		return &dto.PayOrderResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    "failed to register order, contact the administrator",
			MsgLog:     "failed to register order",
		}, err
	}

	return &dto.PayOrderResponse{
		StatusCode: http.StatusOK,
		MsgUser:    "order registered",
		MsgLog:     "getting balance for user successfully",
	}, nil
}

func (m *DefAccrualManager) GetOrders(req *dto.GetOrdersRequest) (*dto.GetOrdersResponse, error) {
	orders, err := m.storage.Orders().GetOrdersByUserID(req.UserID)
	if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
		return &dto.GetOrdersResponse{
			Orders:     nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get orders",
		}, err
	}

	if len(*orders) == 0 {
		msg := "there is no order"
		return &dto.GetOrdersResponse{
			Orders:     orders,
			StatusCode: http.StatusNoContent,
			MsgUser:    msg,
			MsgLog:     msg,
		}, nil
	}

	return &dto.GetOrdersResponse{
		Orders:     orders,
		StatusCode: http.StatusOK,
		MsgUser:    "",
		MsgLog:     "getting list of registered orders successfully",
	}, nil
}
