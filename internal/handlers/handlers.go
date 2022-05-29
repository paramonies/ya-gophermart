package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

func Register(db *store.PostgresDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "register handler", "request URL", r.URL, "method", r.Method)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to read request body", err, http.StatusInternalServerError)
			return
		}

		var reqBodyJSON struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		err = json.Unmarshal(b, &reqBodyJSON)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", reqBodyJSON.Login, "password", reqBodyJSON.Password)
		hash, err := utils.EncryptPassword(reqBodyJSON.Password)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to encrypt password", err, http.StatusInternalServerError)
			return
		}

		err = db.CreateUser(reqBodyJSON.Login, hash)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to create user in db", err, http.StatusConflict)
			return
		}

		utils.VerifyToken(w, r, hash)

		msg := "user created and registered"
		utils.WriteMsgAsJSON(w, msg, http.StatusOK)
		log.Debug(context.Background(), msg)
	}
}

func Login(db *store.PostgresDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "login handler", "request URL", r.URL, "method", r.Method)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to read request body", err, http.StatusInternalServerError)
			return
		}

		var reqBodyJSON struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		err = json.Unmarshal(b, &reqBodyJSON)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", reqBodyJSON.Login, "password", reqBodyJSON.Password)
		hash, err := utils.EncryptPassword(reqBodyJSON.Password)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to encrypt password", err, http.StatusInternalServerError)
			return
		}

		hashedPassword, err := db.GetHashedPassword(reqBodyJSON.Login)
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to get user from db", err, http.StatusUnauthorized)
			return
		}

		err = utils.VerifyPassword(*hashedPassword, reqBodyJSON.Password)
		if err != nil {
			utils.WriteErrorAsJSON(w, fmt.Sprintf("invalid password: %v", err), errors.New("invalid password"), http.StatusUnauthorized)
			return
		}

		utils.VerifyToken(w, r, hash)

		msg := "user is authorized"
		utils.WriteMsgAsJSON(w, msg, http.StatusOK)
		log.Debug(context.Background(), msg)
	})
}

func CreateOrder(db *store.PostgresDB, ac *provider.AccrualClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "create order handler", "request URL", r.URL, "method", r.Method)

		orderNumberBin, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to read order number", err, http.StatusInternalServerError)
			return
		}

		orderNumber, err := strconv.Atoi(string(orderNumberBin))
		if err != nil {
			utils.WriteErrorAsJSON(w, "order number is not integer", err, http.StatusInternalServerError)
			return
		}

		if utils.Valid(orderNumber) {
			msg := "order number is invalid according to Luhn"
			utils.WriteErrorAsJSON(w, msg, errors.New(msg), http.StatusInternalServerError)
			return
		}

		cookie, err := r.Cookie("id")
		if err != nil {
			utils.WriteErrorAsJSON(w, "failed to get cookie", err, http.StatusInternalServerError)
			return
		}

		userID := cookie.Value
		err = db.CreateOrder(orderNumber, userID)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyCreatedByUser) {
				utils.WriteErrorAsJSON(w, err.Error(), err, http.StatusOK)
				return
			}

			if errors.Is(err, store.ErrAlreadyCreatedByOtherUser) {
				utils.WriteErrorAsJSON(w, err.Error(), err, http.StatusConflict)
				return
			}

			utils.WriteErrorAsJSON(w, "failed to create order", err, http.StatusInternalServerError)
			return
		}

		go UpdateOrder(db, ac, orderNumber, userID)

		msg := "order accepted for processing"
		utils.WriteMsgAsJSON(w, msg, http.StatusAccepted)
		log.Debug(context.Background(), msg)
	}
}

func SelectOrders(db *store.PostgresDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "select orders handler", "request URL", r.URL, "method", r.Method)

		cookie, err := r.Cookie("token")
		if err != nil {
			utils.WriteErrorAsJSON(w, fmt.Sprintf("failed to get cookie: %v", err), errors.New("unauthorized"), http.StatusUnauthorized)
			return
		}

		orders, err := db.SelectOrders(cookie.Value)
		if err != nil {
			if errors.Is(err, store.ErrOrdersNotFound) {
				utils.WriteErrorAsJSON(w, "orders not found", err, http.StatusNoContent)
				return
			}

			utils.WriteErrorAsJSON(w, "failed to get orders", err, http.StatusInternalServerError)
			return
		}

		utils.WriteResponseAsJSON(w, orders, http.StatusOK)
		log.Debug(context.Background(), "getting list of orders successfully")
	}
}

func UpdateOrder(db *store.PostgresDB, ac *provider.AccrualClient, orderNumber int, userID string) {
	order, err := ac.GetOrder(orderNumber)
	if err != nil {
		log.Error(context.Background(), "failed to get order from provider", err)
		return
	}

	err = db.UpdateOrder(*order)
	if err != nil {
		log.Error(context.Background(), "failed to update order in db", err)
		return
	}

	log.Info(context.Background(), "order updated successfully", "orderNumber", orderNumber, "userID", userID)
}
