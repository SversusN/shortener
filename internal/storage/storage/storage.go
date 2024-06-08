package storage

type Storage interface {
	GetURL(id string) (string, error)
	SetURL(id string, targetURL string) error
	GetKey(url string) (string, error)
	SetURLBatch(u map[string]string) error
}

type Pinger interface {
	Ping() error
}
