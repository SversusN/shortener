package primitivestorage

import (
	"errors"
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

func (m *MapStorage) SetURL(id string, originalURL string) error {
	_, loaded := m.data.LoadOrStore(id, originalURL)
	if loaded {
		log.Println("key is already in the storage")
	} else {
		if m.helper != nil {
			m.helper.WriteFile(lenSyncMap(m.data), originalURL, id)
		}
	}
	return nil
}

func (m *MapStorage) SetURLBatch(u map[string]string) error {

	for s := range u {
		_, ok := m.data.LoadOrStore(s, u[s])
		if ok {
			log.Println("key is already in the storage")
		} else {
			if m.helper != nil {
				m.helper.WriteFile(lenSyncMap(m.data), s, u[s])
			}
		}
	}
	return nil
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
		ok = false
		return false
	})
	if ok {
		return storedKey, nil
	}
	return "", errors.New("key not found")
}

func lenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}
