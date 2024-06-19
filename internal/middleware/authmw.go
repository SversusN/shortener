package middleware

import (
	"context"
	"fmt"
	"github.com/SversusN/shortener/internal/storage/storage"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	TokenExp   = time.Minute * 180
	SecretKey  = "secret"
	NameCookie = "Token"
	CtxUser    = "UserID"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}
type AuthMW struct {
	db  storage.Storage
	ctx *context.Context
}

func NewAuthMW(db storage.Storage, ctx *context.Context) *AuthMW {
	return &AuthMW{db: db, ctx: ctx}
}

func BuildNewToken(userID int) (string, error) {
	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
	},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("building new jwt: %w", err)
	}
	return stringToken, nil
}

func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return -1, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return -1, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

func (a AuthMW) AuthMWfunc(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(NameCookie)
		if err == nil {
			userID, err := GetUserID(cookie.Value)
			if userID != -1 && err == nil {
				ctx := context.WithValue(r.Context(), CtxUser, userID)
				h.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		saver, ok := a.db.(storage.UserStorage)
		if !ok {
			return
		}
		userID, err := saver.CreateUser(r.Context())
		if err != nil {
			fmt.Errorf("err while creating new user in auth mw: %v", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		token, err := BuildNewToken(userID)
		if err != nil {
			fmt.Errorf("err while creating new jwt: %v", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  NameCookie,
			Value: token,
		})
		ctx := context.WithValue(*a.ctx, CtxUser, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
