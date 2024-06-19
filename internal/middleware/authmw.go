package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/SversusN/shortener/internal/storage/storage"
	"log"
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
	db storage.Storage
}

func NewAuthMW(db storage.Storage) *AuthMW {
	return &AuthMW{db: db}
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
		return -1, errors.New("error parsing token")
	}

	if !token.Valid {
		return -1, errors.New("invalid token")
	}

	return claims.UserID, nil
}

func (a AuthMW) AuthMWfunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(NameCookie)
		if err == nil {
			userID, err := GetUserID(cookie.Value)
			if userID != -1 && err == nil {
				ctx := context.WithValue(r.Context(), CtxUser, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("pls, clear cookie data"))
			}
		}
		saver, ok := a.db.(storage.UserStorage)
		if !ok {
			return
		}
		userID, err := saver.CreateUser(r.Context())
		if err != nil {
			log.Print("err while creating new user in auth mw")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		token, err := BuildNewToken(userID)
		if err != nil {
			log.Print("err while building new token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  NameCookie,
			Value: token,
		})
		ctx := context.WithValue(r.Context(), CtxUser, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
