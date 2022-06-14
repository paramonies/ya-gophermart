package server

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/paramonies/ya-gophermart/internal/server/dto"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

type OrderManager interface {
	Load(req *dto.LoadOrderRequest) (*dto.LoadOrderResponse, error)
	ListProcessed(req *dto.ListOrdersRequest) (*dto.ListOrdersResponse, error)
	GetBalance(req *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error)
}

type OrderHandler struct {
	manager OrderManager
}

func NewOrderHandler(manager OrderManager) *OrderHandler {
	return &OrderHandler{
		manager: manager,
	}
}

func (h *OrderHandler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		orderNumberBin, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			msg := "failed to read order number"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusBadRequest)
			return
		}

		orderNumber, err := strconv.Atoi(string(orderNumberBin))
		if err != nil {
			msg := "order number is not integer"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusBadRequest)
			return
		}
		log.Debug(context.Background(), "oder info", "orderNumber", orderNumber)

		orderInfo := &dto.LoadOrderRequest{
			User:        user,
			OrderNumber: orderNumber,
		}
		resp, err := h.manager.Load(orderInfo)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgLog)
	}
}

func (h *OrderHandler) ListProcessed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		listOrderInfo := &dto.ListOrdersRequest{
			UserID: user.ID,
		}

		resp, err := h.manager.ListProcessed(listOrderInfo)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		if len(*resp.Orders) == 0 {
			utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
			return
		}

		orders := make([]Order, 0)
		for _, item := range *resp.Orders {
			o := Order{
				Number:     item.OrderNumber,
				Status:     item.Status,
				Accrual:    item.Accrual,
				UploadedAt: item.UpdatedAt.Format(time.RFC3339),
			}
			orders = append(orders, o)
		}
		utils.WriteResponseAsJSON(w, orders, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgUser)
	}
}

func (h *OrderHandler) GetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		getBalanceInfo := &dto.GetBalanceRequest{
			UserID: user.ID,
		}

		resp, err := h.manager.GetBalance(getBalanceInfo)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		res := Balance{
			Current:   *resp.CurrentSum,
			Withdrawn: *resp.WithdrawnSum,
		}

		utils.WriteResponseAsJSON(w, res, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgLog)
	}
}
