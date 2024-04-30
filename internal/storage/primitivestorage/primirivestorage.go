package primitivestorage

import (
	"errors"
)

type MapStorage struct {
	data map[string]string
}

func NewStorage() *MapStorage {
	return &MapStorage{
		data: make(map[string]string),
	}
}

func (m MapStorage) GetURL(id string) (string, error) {
	originalURL, ok := m.data[id]
	if !ok {
		return "", errors.New("original url not found")
	}
	return originalURL, nil
}

func (m MapStorage) SetURL(id string, originalURL string) error {
	if _, ok := m.data[id]; ok {
		return errors.New("short-key already created")
	}
	m.data[id] = originalURL
	return nil
}
