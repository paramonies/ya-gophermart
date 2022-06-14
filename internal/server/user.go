package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/server/dto"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

type UserManager interface {
	Create(req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	Login(req *dto.LoginUserRequest) (*dto.LoginUserResponse, error)
}

type UserHandler struct {
	manager UserManager
}

func NewUserHandler(manager UserManager) *UserHandler {
	return &UserHandler{
		manager: manager,
	}
}

func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := "failed to register user"

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", req.Login, "password", req.Password)

		userReq := &dto.CreateUserRequest{
			Login:    req.Login,
			Password: req.Password,
		}
		resp, err := h.manager.Create(userReq)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		utils.SetCookie(w, *resp.Token)
		utils.WriteMsgAsJSON(w, resp.MsgUser, resp.StatusCode)
		log.Debug(context.Background(), resp.MsgLog)
	}
}

func (h *UserHandler) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := "failed to sign user in"

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteErrorAsJSON(w, msg, "failed to unmarshal request body", err, http.StatusBadRequest)
			return
		}

		log.Debug(context.Background(), "user credentials", "login", req.Login, "password", req.Password)

		userReq := &dto.LoginUserRequest{
			Login:    req.Login,
			Password: req.Password,
		}
		resp, err := h.manager.Login(userReq)
		if err != nil {
			utils.WriteErrorAsJSON(w, resp.MsgUser, resp.MsgLog, err, resp.StatusCode)
			return
		}

		utils.SetCookie(w, *resp.Token)
		utils.WriteMsgAsJSON(w, resp.MsgUser, http.StatusOK)
		log.Debug(context.Background(), resp.MsgLog)
	})
}
