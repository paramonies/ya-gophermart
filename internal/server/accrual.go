package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/paramonies/ya-gophermart/internal/server/dto"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

type AccrualManager interface {
	PayOrder(req *dto.PayOrderRequest) (*dto.PayOrderResponse, error)
	GetOrders(req *dto.GetOrdersRequest) (*dto.GetOrdersResponse, error)
}

type AccrualHandler struct {
	manager AccrualManager
}

func NewAccrualHandler(manager AccrualManager) *AccrualHandler {
	return &AccrualHandler{
		manager: manager,
	}
}

func (h *AccrualHandler) PayOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		var req PayOrderRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteErrorAsJSON(w, "invalid format", "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		orderNumber, err := strconv.Atoi(string(req.OrderNumber))
		if err != nil {
			msg := "order number is not integer"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusBadRequest)
			return
		}
		log.Debug(context.Background(), "oder info", "orderNumber", orderNumber)

		payOrderReq := &dto.PayOrderRequest{
			UserID:      user.ID,
			OrderNumber: orderNumber,
			Price:       req.Price,
		}
		resp, err := h.manager.PayOrder(payOrderReq)
		if err != nil {
			if resp.StatusCode == http.StatusOK {
				utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
				return
			}
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgLog)
	}
}

func (h *AccrualHandler) GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		getOrdersReq := &dto.GetOrdersRequest{
			UserID: user.ID,
		}

		resp, err := h.manager.GetOrders(getOrdersReq)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		if len(*resp.Orders) == 0 {
			utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
			return
		}

		orders := make([]OrderResponse, 0)
		for _, item := range *resp.Orders {
			o := OrderResponse{
				OrderNumber: item.OrderNumber,
				Price:       item.Price,
				UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
			}
			orders = append(orders, o)
		}

		utils.WriteResponseAsJSON(w, orders, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgLog)
	}
}
