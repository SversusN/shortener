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

func NewStorage(helper utils.FileHelper) *MapStorage {
	dirtyMap := helper.ReadFile()
	return &MapStorage{
		data:   dirtyMap,
		helper: &helper,
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
		m.helper.WriteFile(originalURL, id)
	}
	return nil
}
