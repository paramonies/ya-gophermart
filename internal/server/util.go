package server

import (
	"context"
	"errors"

	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

func getUser(ctx context.Context) (*dto.User, error) {
	u, ok := ctx.Value(User).(*dto.User)
	if !ok {
		return nil, errors.New("failed to get user from context")
	}

	return u, nil
}
