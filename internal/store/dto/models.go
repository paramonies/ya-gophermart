package dto

import "time"

type User struct {
	ID           string
	Login        string
	PasswordHash string
	Token        string
}

type OrderAccrual struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	Accrual     float64   `json:"accrual,omitempty"`
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at" format:"RFC3339"`
}

type ProviderOrder struct {
	Number  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Order struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	Price       float64   `json:"price,omitempty"`
	UserID      string    `json:"user_id"`
	UpdatedAt   time.Time `json:"updated_at" format:"RFC3339"`
}
