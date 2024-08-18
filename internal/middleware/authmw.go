package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/SversusN/shortener/internal/storage/storage"
)

const (
	TokenExp           = time.Minute * 180 //Время жизни токена
	SecretKey          = "secret"          // секретный ключ
	NameCookie         = "Token"           // наименование куки в запросе
	CtxUser    ctxUser = "UserID"          //поле ид пользователя в куки
)

type ctxUser string

// Claims тиа для указания UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// AuthMW структура middleware авторизации
type AuthMW struct {
	db storage.Storage
}

// NewAuthMW конструктор объекта авторизации
func NewAuthMW() *AuthMW {
	return &AuthMW{}
}

// BuildNewToken функция генерации токена новому пользователю
func BuildNewToken(userID string) (string, error) {
	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
	},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", errors.New("error signing token")
	}
	return stringToken, nil
}

// GetUserID получение ИД пользователя из токена
func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return "", errors.New("error parsing token")
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

// AuthMWfunc Функция аутентифкации
func (a AuthMW) AuthMWfunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(NameCookie)
		if err == nil {
			userID, err := GetUserID(cookie.Value)
			if userID != "" && err == nil {
				ctx := context.WithValue(r.Context(), CtxUser, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("pls, clear cookie data"))
				return
			}
		}
		//ИД пользователя если обращение происходит первый раз (uuid)
		userID := uuid.NewString()
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
