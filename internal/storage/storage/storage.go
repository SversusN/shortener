package storage

type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string) error
}

type Pinger interface {
	Ping() error
}
