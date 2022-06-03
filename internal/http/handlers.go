package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	pgxv4 "github.com/jackc/pgx/v4"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/store/dto"
	"github.com/paramonies/ya-gophermart/internal/store/pgx"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

func Register(storage store.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := "failed to register user"

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", req.Login, "password", req.Password)

		user, err := storage.Users().GetByName(req.Login)
		if user != nil {
			msg = "user have already created"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusConflict)
			return
		}

		if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
			utils.WriteErrorAsJSON(w, "Ooops!", "failed to get user from db", err, http.StatusInternalServerError)
			return
		}

		hash, err := utils.EncryptPassword(req.Password)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to encrypt password", err, http.StatusInternalServerError)
			return
		}

		err = storage.Users().Create(req.Login, hash)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to create user in db", err, http.StatusInternalServerError)
			return
		}

		token := utils.GenerateToken()
		err = storage.Users().SetToken(req.Login, token)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to create user in db", err, http.StatusInternalServerError)
			return
		}

		utils.SetCookie(w, token)

		msgOK := "user created and registered"
		utils.WriteMsgAsJSON(w, msgOK, http.StatusOK)
		log.Debug(context.Background(), msgOK)
	}
}

func Login(storage store.Connector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := "failed to sign user in"

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", req.Login, "password", req.Password)

		user, err := storage.Users().GetByName(req.Login)
		if err != nil {
			if errors.Is(err, pgx.ErrConstraintViolation) {
				msg = err.Error()
			}
			utils.WriteErrorAsJSON(w, msg, "failed to get user from db", err, http.StatusConflict)
			return
		}

		err = utils.VerifyPassword(user.PasswordHash, req.Password)
		if err != nil {
			msg = "invalid password"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusUnauthorized)
			return
		}

		token := utils.GenerateToken()
		err = storage.Users().SetToken(req.Login, token)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to set token in db", err, http.StatusInternalServerError)
			return
		}

		utils.SetCookie(w, token)

		msgOK := "user is authorized"
		utils.WriteMsgAsJSON(w, msgOK, http.StatusOK)
		log.Debug(context.Background(), msgOK)
	})
}

func getUser(ctx context.Context) (*dto.User, error) {
	u, ok := ctx.Value(User).(*dto.User)
	if !ok {
		return nil, errors.New("failed to get user from context")
	}

	return u, nil
}

func LoadOrder(storage store.Connector, ac *provider.AccrualClient) http.HandlerFunc {
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

		if !utils.Valid(orderNumber) {
			msg := "order number is invalid according to Luhn"
			utils.WriteErrorAsJSON(w, msg, msg, errors.New(msg), http.StatusUnprocessableEntity)
			return
		}

		order, err := storage.Accruals().GetOrderByOrderNumber(orderNumber)
		if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get order from db", err, http.StatusInternalServerError)
			return
		}

		if err == nil && order != nil && order.UserID != user.ID {
			msg := "order was loaded by other user"
			utils.WriteErrorAsJSON(w, msg, msg, nil, http.StatusConflict)
			return
		}

		if err == nil && order != nil && order.UserID == user.ID {
			utils.WriteMsgAsJSON(w, "order already loaded", http.StatusOK)
			return
		}

		err = storage.Accruals().LoadOrder(orderNumber, user.ID)
		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to load order to accrual table", err, http.StatusInternalServerError)
			return

		}

		utils.WriteMsgAsJSON(w, "order load to processing", http.StatusAccepted)
		log.Debug(context.Background(), "order load to accrual table")
	}
}

func ListProcessedOrders(storage store.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		list, err := storage.Accruals().GetOrderByUserID(user.ID)
		if len(*list) == 0 {
			utils.WriteMsgAsJSON(w, "there is no order", http.StatusNoContent)
			return
		}

		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get orders from accrual table", err, http.StatusInternalServerError)
			return
		}

		orders := make([]Order, 0)
		for _, item := range *list {
			o := Order{
				Number:     item.OrderNumber,
				Status:     item.Status,
				Accrual:    item.Accrual,
				UploadedAt: item.UpdatedAt.Format(time.RFC3339),
			}
			orders = append(orders, o)
		}
		utils.WriteResponseAsJSON(w, orders, http.StatusOK)
		log.Debug(context.Background(), "getting list of loaded orders successfully")
	}
}

func GetBalance(storage store.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		totalAccrual, err := storage.Accruals().GetOrdersAccrualForUser(user.ID)
		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get total accruals for all orders for user", err, http.StatusInternalServerError)
			return
		}

		totalPrice, err := storage.Orders().GetOrdersPriceForUser(user.ID)
		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get total prices for all orders for user", err, http.StatusInternalServerError)
			return
		}

		res := Balance{
			Current:   *totalAccrual - *totalPrice,
			Withdrawn: *totalPrice,
		}

		utils.WriteResponseAsJSON(w, res, http.StatusOK)
		log.Debug(context.Background(), "getting balance for user successfully")
	}
}

func PayOrder(storage store.Connector) http.HandlerFunc {
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

		if !utils.Valid(orderNumber) {
			msg := "order number is invalid according to Luhn"
			utils.WriteErrorAsJSON(w, msg, msg, errors.New(msg), http.StatusUnprocessableEntity)
			return
		}

		totalAccrual, err := storage.Accruals().GetOrdersAccrualForUser(user.ID)
		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get total accruals for all orders for user", err, http.StatusInternalServerError)
			return
		}

		totalPrice, err := storage.Orders().GetOrdersPriceForUser(user.ID)
		if err != nil {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get total prices for all orders for user", err, http.StatusInternalServerError)
			return
		}

		balance := *totalAccrual - *totalPrice
		if balance <= 0 {
			utils.WriteErrorAsJSON(w, "not enought money on your account", "negative balance for user", err, http.StatusPaymentRequired)
			return
		}

		err = storage.Orders().Register(user.ID, req.OrderNumber, float64(req.Price))
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to register order, contact the administrator", "failed to register order", err, http.StatusInternalServerError)
			return
		}

		utils.WriteMsgAsJSON(w, "order registered", http.StatusOK)
		log.Debug(context.Background(), "getting balance for user successfully")
	}
}

func GetOrders(storage store.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r.Context())
		if err != nil {
			utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
			return
		}

		list, err := storage.Orders().GetOrdersByUserID(user.ID)
		if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
			utils.WriteErrorAsJSON(w, "oops)", "failed to get orders", err, http.StatusUnauthorized)
			return
		}

		if len(*list) == 0 {
			utils.WriteMsgAsJSON(w, "there is no order", http.StatusNoContent)
			return
		}

		orders := make([]OrderResponse, 0)
		for _, item := range *list {
			o := OrderResponse{
				OrderNumber: item.OrderNumber,
				Price:       item.Price,
				UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
			}

			orders = append(orders, o)
		}

		utils.WriteResponseAsJSON(w, orders, http.StatusOK)
		log.Debug(context.Background(), "getting list of registered orders successfully")
	}
}
