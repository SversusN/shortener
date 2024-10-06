// интерфейс storage для работы с хранилищем
package storage

import (
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
	"sync"
)

// Storage интерфейс описания методов хранилища
type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string, userID string) (string, error)
	SetURLBatch(u map[string]entity.UserURL) (map[string]entity.UserURL, error)
	GetUserUrls(userID string) (any, error)
	DeleteUserURLs(userID string, group *sync.WaitGroup) (chan string, error)
	GetStats() (int, int, error)
}

// Pinger интерфейс для проверки соединения PostgreSQL
type Pinger interface {
	Ping() error
}
