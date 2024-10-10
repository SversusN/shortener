package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключ с id пользователя.
const UserIDMetaKey userID = "user_id"

type userID string

// AuthInterceptor описывает структуру интерцептора аутентификации
type AuthInterceptor struct {
}

// NewAuthInterceptor создает аутентификатор
func NewAuthInterceptor(ctx context.Context) *AuthInterceptor {
	return &AuthInterceptor{}
}

// AuthenticateUser идентифицирует пользователя в запросе.
func (i *AuthInterceptor) AuthenticateUser(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	var ID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("user_id")
		if len(values) > 0 {
			idString := values[0]
			ID = idString
		} else {
			return nil, status.Error(codes.Unauthenticated, "wrong user id format")
		}
		ctx = context.WithValue(ctx, UserIDMetaKey, ID)
	}
	return handler(ctx, req)
}
