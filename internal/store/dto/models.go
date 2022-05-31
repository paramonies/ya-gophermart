package dto

import "time"

type User struct {
	ID           string
	Login        string
	PasswordHash string
	Token        string
}

type Order struct {
	ID        string    `json:"id"`
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	Accural   float64   `json:"accrual,omitempty"`
	UpdatedAt time.Time `json:"updated_at" format:"RFC3339"`
}

type ProviderOrder struct {
	Number  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
