package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/paramonies/ya-gophermart/pkg/log"
)

func WriteErrorAsJSON(w http.ResponseWriter, msgUser, msgLog string, err error, code int) {
	log.Error(context.Background(), msgLog, err)

	respErr := struct {
		Err string `json:"error"`
	}{
		Err: msgUser,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	respJSON, e := json.Marshal(respErr)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	_, e = w.Write(respJSON)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func WriteMsgAsJSON(w http.ResponseWriter, msg string, code int) {
	respMsg := struct {
		Msg string `json:"message"`
	}{
		Msg: msg,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	respJSON, err := json.Marshal(respMsg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(respJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteResponseAsJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	respJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(respJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func VerifyToken(w http.ResponseWriter, r *http.Request, hash string) {
	_, err := r.Cookie("token")
	if errors.Is(err, http.ErrNoCookie) {
		cookie := &http.Cookie{
			Name:   "token",
			Value:  hash,
			Secure: false,
		}
		http.SetCookie(w, cookie)
		log.Debug(context.Background(), "set cookie for user")
	}
}
