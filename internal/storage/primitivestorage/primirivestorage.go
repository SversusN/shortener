package primitivestorage

import (
	"errors"
	"github.com/SversusN/shortener/internal/internalerrors"
	"log"
	"sync"

	"github.com/SversusN/shortener/internal/pkg/utils"
)

type MapStorage struct {
	data   *sync.Map
	helper *utils.FileHelper
}

// NewStorage хелпер межет придти nil
func NewStorage(helper *utils.FileHelper, err error) *MapStorage {
	if err != nil {
		return &MapStorage{
			data: &sync.Map{},
		}
	}
	tempMap := helper.ReadFile()
	return &MapStorage{
		data:   tempMap,
		helper: helper,
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

func (m *MapStorage) SetURL(shortURL string, originalURL string) (string, error) {
	result, err := m.GetKey(originalURL)
	switch {
	case errors.Is(err, internalerrors.ErrNotFound):
		{
			_, loaded := m.data.LoadOrStore(shortURL, originalURL)
			if m.helper != nil {
				m.helper.WriteFile(lenSyncMap(m.data), originalURL, shortURL)
			}
			if loaded {
				log.Println("key is already in the storage")
			}
			return shortURL, nil
		}
	case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
		{
			return result, internalerrors.ErrOriginalURLAlreadyExists
		}
	default:
		{
			return "", err
		}
	}
}

func (m *MapStorage) SetURLBatch(u map[string]string) (map[string]string, error) {
	returned := make(map[string]string)
	var possibleDoubleError error
	for s := range u {
		result, err := m.GetKey(u[s])
		switch {
		case errors.Is(err, internalerrors.ErrNotFound):
			{
				_, loaded := m.data.LoadOrStore(s, u[s])
				if m.helper != nil {
					m.helper.WriteFile(lenSyncMap(m.data), s, u[s])
				}
				if loaded {
					log.Println("key is already in the storage")
				}
				returned[s] = u[s]
			}
		case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
			{
				possibleDoubleError = err
				returned[s] = result
			}
		default:
			{
				return returned, err
			}
		}
	}
	return returned, possibleDoubleError
}

func (m *MapStorage) GetKey(url string) (string, error) {
	var storedKey string
	var ok bool
	m.data.Range(func(key, value interface{}) bool {
		if value == url {
			storedKey = key.(string)
			ok = true
			return true
		}
		return false
	})
	if ok {
		return storedKey, internalerrors.ErrOriginalURLAlreadyExists
	} else {
		return "", internalerrors.ErrNotFound
	}
}

func lenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}
