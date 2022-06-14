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

type DefUserManager struct {
	storage store.Connector
}

func NewUserManager(storage store.Connector) managers.UserManager {
	return &DefUserManager{
		storage: storage,
	}
}

func (m *DefUserManager) Create(req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	user, err := m.storage.Users().GetByName(req.Login)
	if user != nil {
		msg := "user have already created"
		return &dto.CreateUserResponse{
			Token:      nil,
			StatusCode: http.StatusConflict,
			MsgUser:    msg,
			MsgLog:     msg,
		}, errors.New(msg)
	}

	if err != nil && !errors.Is(err, pgxv4.ErrNoRows) {
		return &dto.CreateUserResponse{
			Token:      nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgOoops,
			MsgLog:     "failed to get user from db",
		}, err
	}

	hash, err := utils.EncryptPassword(req.Password)
	if err != nil {
		return &dto.CreateUserResponse{
			Token:      nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgRegisterErr,
			MsgLog:     "failed to encrypt password",
		}, err
	}

	err = m.storage.Users().Create(req.Login, hash)
	if err != nil {
		return &dto.CreateUserResponse{
			Token:      nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgRegisterErr,
			MsgLog:     "failed to create user in db",
		}, err
	}

	token := utils.GenerateToken()
	err = m.storage.Users().SetToken(req.Login, token)
	if err != nil {
		return &dto.CreateUserResponse{
			Token:      nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    msgRegisterErr,
			MsgLog:     "failed to create user in db",
		}, err
	}

	return &dto.CreateUserResponse{
		Token:      &token,
		StatusCode: http.StatusOK,
		MsgUser:    msgRegisterOK,
		MsgLog:     msgRegisterOK,
	}, nil
}

func (m *DefUserManager) Login(req *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	user, err := m.storage.Users().GetByName(req.Login)
	if err != nil {
		var msg string
		if errors.Is(err, pgxv4.ErrNoRows) {
			msg = "you are not registered yet"
		}
		return &dto.LoginUserResponse{
			Token:      nil,
			StatusCode: http.StatusUnauthorized,
			MsgUser:    msg,
			MsgLog:     "failed to get user from db",
		}, err
	}

	err = utils.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		msg := "invalid password"
		return &dto.LoginUserResponse{
			Token:      nil,
			StatusCode: http.StatusUnauthorized,
			MsgUser:    msg,
			MsgLog:     "failed to get user from db",
		}, err
	}

	token := utils.GenerateToken()
	err = m.storage.Users().SetToken(req.Login, token)
	if err != nil {
		return &dto.LoginUserResponse{
			Token:      nil,
			StatusCode: http.StatusInternalServerError,
			MsgUser:    "failed to sign user in",
			MsgLog:     "failed to set token in db",
		}, err
	}

	return &dto.LoginUserResponse{
		Token:      &token,
		StatusCode: http.StatusOK,
		MsgUser:    msgAuthorizedOK,
		MsgLog:     msgAuthorizedOK,
	}, nil
}
