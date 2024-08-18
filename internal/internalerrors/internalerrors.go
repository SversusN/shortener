// Пакет описания внутренних ошибок приложения
package internalerrors

import (
	"errors"
	"fmt"
)

var (
	ErrOriginalURLAlreadyExists = errors.New("original url already exists") // Оригинальный URL есть в хранилище
	ErrKeyAlreadyExists         = errors.New("key already exists")          //BUG(SversusN) Сокращенный ключ был выдан. Коллизия астрономически мала
	ErrNotFound                 = errors.New("key not found")               // Ключ не найден
	ErrUserTypeError            = errors.New("user type error")             // Ошибка получения ИД пользователя
	ErrUserNotFound             = errors.New("user not found error")        // Ошибка наличия пользователя
	ErrDeleted                  = errors.New("try get deleted error")       //Попытка получения удаленной ссылки
)

// ConflictError тип внутренней ошибки конфликта
type ConflictError struct {
	Err      error
	ShortURL string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict: short URL already exists: %v", e.ShortURL)
}

func NewConflictError(err error, text string) error {
	return &ConflictError{Err: err, ShortURL: text}
}
