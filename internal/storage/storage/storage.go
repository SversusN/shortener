package storage

import (
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
)

type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string, userID string) (string, error)
	SetURLBatch(u map[string]entity.UserURL) (map[string]entity.UserURL, error)
	GetUserUrls(userID string) (any, error)
	DeleteUserURLs(userID string) (chan string, error)
}

type Pinger interface {
	Ping() error
}
