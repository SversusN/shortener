package storage

import "context"

type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string) error
<<<<<<< Updated upstream
=======
	GetKey(url string) (string, error)
	SetUrlBatch(c context.Context, u map[string]string) error
}

type Pinger interface {
	Ping() error
>>>>>>> Stashed changes
}
