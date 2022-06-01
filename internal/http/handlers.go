package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

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

func CreateOrder(storage store.Connector, ac *provider.AccrualClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUser(r.Context())
		if err != nil {
			msg := "failed to get user from context"
			utils.WriteErrorAsJSON(w, msg, msg, err, http.StatusUnauthorized)
		}

		//fmt.Println("!!! %s", user.Login)
		//		orderNumberBin, err := io.ReadAll(r.Body)
		//		defer r.Body.Close()
		//		if err != nil {
		//			utils.WriteErrorAsJSON(w, "failed to read order number", err, http.StatusInternalServerError)
		//			return
		//		}
		//
		//		orderNumber, err := strconv.Atoi(string(orderNumberBin))
		//		if err != nil {
		//			utils.WriteErrorAsJSON(w, "order number is not integer", err, http.StatusInternalServerError)
		//			return
		//		}
		//
		//		if utils.Valid(orderNumber) {
		//			msg := "order number is invalid according to Luhn"
		//			utils.WriteErrorAsJSON(w, msg, errors.New(msg), http.StatusInternalServerError)
		//			return
		//		}
		//
		//		cookie, err := r.Cookie("id")
		//		if err != nil {
		//			utils.WriteErrorAsJSON(w, "failed to get cookie", err, http.StatusInternalServerError)
		//			return
		//		}
		//
		//		userID := cookie.Value
		//		err = storage.Orders().CreateOrder(orderNumber, userID)
		//		if err != nil {
		//			if errors.Is(err, pgx.ErrAlreadyCreatedByUser) {
		//				utils.WriteErrorAsJSON(w, err.Error(), err, http.StatusOK)
		//				return
		//			}
		//
		//			if errors.Is(err, pgx.ErrAlreadyCreatedByOtherUser) {
		//				utils.WriteErrorAsJSON(w, err.Error(), err, http.StatusConflict)
		//				return
		//			}
		//
		//			utils.WriteErrorAsJSON(w, "failed to create order", err, http.StatusInternalServerError)
		//			return
		//		}
		//
		//		go UpdateOrder(storage, ac, orderNumber, userID)
		//
		//		msg := "order accepted for processing"
		//		utils.WriteMsgAsJSON(w, msg, http.StatusAccepted)
		//		log.Debug(context.Background(), msg)
	}
}

//
//func SelectOrders(storage store.Connector) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Debug(context.Background(), "select orders handler", "request URL", r.URL, "method", r.Method)
//
//		cookie, err := r.Cookie("token")
//		if err != nil {
//			utils.WriteErrorAsJSON(w, fmt.Sprintf("failed to get cookie: %v", err), errors.New("unauthorized"), http.StatusUnauthorized)
//			return
//		}
//
//		orders, err := storage.Orders().SelectOrders(cookie.Value)
//		if err != nil {
//			if errors.Is(err, pgx.ErrOrdersNotFound) {
//				utils.WriteErrorAsJSON(w, "orders not found", err, http.StatusNoContent)
//				return
//			}
//
//			utils.WriteErrorAsJSON(w, "failed to get orders", err, http.StatusInternalServerError)
//			return
//		}
//
//		utils.WriteResponseAsJSON(w, orders, http.StatusOK)
//		log.Debug(context.Background(), "getting list of orders successfully")
//	}
//}
//
//func UpdateOrder(storage store.Connector, ac *provider.AccrualClient, orderNumber int, userID string) {
//	order, err := ac.GetOrder(orderNumber)
//	if err != nil {
//		log.Error(context.Background(), "failed to get order from provider", err)
//		return
//	}
//
//	err = storage.Orders().UpdateOrder(*order)
//	if err != nil {
//		log.Error(context.Background(), "failed to update order in db", err)
//		return
//	}
//
//	log.Info(context.Background(), "order updated successfully", "orderNumber", orderNumber, "userID", userID)
//}
//
//func GetBalance(storage store.Connector) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Debug(context.Background(), "get balance handler", "request URL", r.URL, "method", r.Method)
//
//		//verify auth
//
//	}
//}
