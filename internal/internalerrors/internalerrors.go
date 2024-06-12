package internalerrors

import (
	"errors"
	"fmt"
)

var (
	ErrOriginalURLAlreadyExists = errors.New("original url already exists")
	ErrKeyAlreadyExists         = errors.New("key already exists")
	ErrNotFound                 = errors.New("key not found")
)

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
