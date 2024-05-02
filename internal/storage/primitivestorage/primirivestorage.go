package primitivestorage

import (
	"errors"
	"log"
	"sync"
)

type MapStorage struct {
	data sync.Map
	//data map[string]string
}

func NewStorage() *MapStorage {
	return &MapStorage{
		data: sync.Map{},
	}
}

func (m *MapStorage) GetURL(id string) (string, error) {
	originalURL, ok := m.data.Load(id)
	if !ok {
		return "", errors.New("original url not found")
	}
	s := originalURL.(string)
	return s, nil
}

func (m *MapStorage) SetURL(id string, originalURL string) error {
	_, loaded := m.data.LoadOrStore(id, originalURL)
	if loaded {
		log.Println("key is already in the storage")
	}
	return nil
}
