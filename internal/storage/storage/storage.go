package storage

import "context"

type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string) (string, error)
	SetURLBatch(u map[string]string) (map[string]string, error)
}

type Pinger interface {
	Ping() error
}

type UserStorage interface {
	GetUserUrls(userID int) (any, error)
	CreateUser(ctx context.Context) (int, error)
	SetUserURL(id string, targetURL string, userID int) (string, error)
	SetUserURLBatch(u map[string]string, userID int) (map[string]string, error)
}
