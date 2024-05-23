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
	//data map[string]string
}

// хелпер межет придти nil
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

func lenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}
