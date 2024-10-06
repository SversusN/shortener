package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключ с id пользователя.
const UserIDMetaKey = "user_id"

// AuthInterceptor описывает структуру интерцептора аутентификации
type AuthInterceptor struct {
}

// NewAuthInterceptor создает аутентификатор
func NewAuthInterceptor(ctx context.Context) *AuthInterceptor {
	return &AuthInterceptor{}
}

// AuthenticateUser идентифицирует пользователя в запросе.
func (i *AuthInterceptor) AuthenticateUser(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error
	var userID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(UserIDMetaKey)
		if len(values) > 0 {
			idString := values[0]
			userID = idString
		}
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "Wrong user id format")
		}
		ctx = context.WithValue(ctx, UserIDMetaKey, userID)
	}
	return handler(ctx, req)
}
