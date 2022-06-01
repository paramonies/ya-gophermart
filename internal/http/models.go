package http

type CtxUser string

var (
	User CtxUser = "user"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
