package app

import (
	"errors"
	"net/http"

	pgxv4 "github.com/jackc/pgx/v4"

	"github.com/paramonies/ya-gophermart/internal/managers"
	"github.com/paramonies/ya-gophermart/internal/server/dto"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/utils"
)

type DefOrderManager struct {
	storage store.Connector
}

func NewOrderManager(storage store.Connector) managers.OrderManager {
	return &DefOrderManager{
		storage: storage,
	}
}

func (m *DefOrderManager) Load(req *dto.LoadOrderRequest) (*dto.LoadOrderResponse, error) {
	if !utils.Valid(req.OrderNumber) {
		msg := "order number is invalid according to Luhn"
		return &dto.LoadOrderResponse{
			StatusCode: http.StatusUnprocessableEntity,
			MsgUser:    msg,
			MsgLog:     msg,
		}, errors.New(msg)
	}

	order, err := m.storage.Accruals().GetOrderByOrderNumber(req.OrderNumber)
	if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
		return &dto.LoadOrderResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get order from db",
		}, err
	}

	if err == nil && order != nil && order.UserID != req.User.ID {
		msg := "order was loaded by other user"
		return &dto.LoadOrderResponse{
			StatusCode: http.StatusConflict,
			MsgUser:    msg,
			MsgLog:     msg,
		}, errors.New(msg)
	}

	if err == nil && order != nil && order.UserID == req.User.ID {
		msg := "order already loaded"
		return &dto.LoadOrderResponse{
			StatusCode: http.StatusOK,
			MsgUser:    msg,
			MsgLog:     msg,
		}, nil
	}

	err = m.storage.Accruals().LoadOrder(req.OrderNumber, req.User.ID)
	if err != nil {
		return &dto.LoadOrderResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to load order to accrual table",
		}, err

	}

	return &dto.LoadOrderResponse{
		StatusCode: http.StatusAccepted,
		MsgUser:    "order load to processing",
		MsgLog:     "order load to accrual table",
	}, err
}

func (m *DefOrderManager) ListProcessed(req *dto.ListOrdersRequest) (*dto.ListOrdersResponse, error) {
	orders, err := m.storage.Accruals().GetOrderByUserID(req.UserID)

	if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
		return &dto.ListOrdersResponse{
			Orders:     nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get orders from accrual table",
		}, err
	}

	if len(*orders) == 0 {
		msg := "there is no order"
		return &dto.ListOrdersResponse{
			Orders:     orders,
			StatusCode: http.StatusNoContent,
			MsgUser:    msg,
			MsgLog:     msg,
		}, nil
	}

	return &dto.ListOrdersResponse{
		Orders:     orders,
		StatusCode: http.StatusOK,
		MsgUser:    "",
		MsgLog:     "getting list of loaded orders successfully",
	}, nil
}

func (m *DefOrderManager) GetBalance(req *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error) {
	totalAccrual, err := m.storage.Accruals().GetOrdersAccrualForUser(req.UserID)
	if err != nil {
		return &dto.GetBalanceResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get total accruals for all orders for user",
		}, err
	}

	totalPrice, err := m.storage.Orders().GetOrdersPriceForUser(req.UserID)
	if err != nil {
		return &dto.GetBalanceResponse{
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get total prices for all orders for user",
		}, err
	}

	CurrentSum := *totalAccrual - *totalPrice
	WithdrawnSum := *totalPrice

	return &dto.GetBalanceResponse{
		CurrentSum:   &CurrentSum,
		WithdrawnSum: &WithdrawnSum,
		StatusCode:   http.StatusOK,
		MsgUser:      "",
		MsgLog:       "getting balance for user successfully",
	}, nil
}
