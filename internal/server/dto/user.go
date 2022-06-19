package dto

type CreateUserRequest struct {
	Login    string
	Password string
}

type CreateUserResponse struct {
	Token      *string
	StatusCode int
	MsgUser    string
	MsgLog     string
}

type LoginUserRequest struct {
	Login    string
	Password string
}

type LoginUserResponse struct {
	Token      *string
	StatusCode int
	MsgUser    string
	MsgLog     string
}
