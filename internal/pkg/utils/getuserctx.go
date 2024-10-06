package utils

import (
	"context"
	"errors"

	"github.com/SversusN/shortener/internal/internalerrors"
)

// GetUserIDFromCtx Получает ИД пользователя из запроса для grpc
func GetUserIDFromCtx(ctx context.Context, value string) (string, error) {
	userID := ctx.Value(value)
	if userID == nil {
		return "", errors.New("user ID is missing")
	}
	userIDInt, ok := userID.(string)
	if !ok {
		return "", internalerrors.ErrUserTypeError
	} else {
		return userIDInt, nil
	}
}
