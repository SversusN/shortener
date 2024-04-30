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
	targetURL, ok := m.data[id]
	if !ok {
		return "", errors.New("original url not found")
	}
	return targetURL, nil
}

func (m MapStorage) SetURL(id string, targetURL string) error {
	if _, ok := m.data[id]; ok {
		return errors.New("short-key already created")
	}
	m.data[id] = targetURL
	return nil
}
