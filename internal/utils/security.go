package utils

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func GenerateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func SetCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(60 * 60 * time.Second),
		Secure:   false,
		HttpOnly: false,
	}

	http.SetCookie(w, cookie)
}
