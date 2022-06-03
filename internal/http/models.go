package http

type CtxUser string

var (
	User CtxUser = "user"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type PayOrderRequest struct {
	OrderNumber string `json:"order"`
	Price       int    `json:"sum"`
}

type OrderResponse struct {
	OrderNumber string  `json:"order"`
	Price       float64 `json:"sum"`
	UpdatedAt   string  `json:"processed_at"`
}
