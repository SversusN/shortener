package primitivestorage

import (
	"errors"
	"github.com/SversusN/shortener/internal/internalerrors"
	"log"
	"sync"

	"github.com/SversusN/shortener/internal/pkg/utils"
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
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
	userURL, ok := m.data.Load(id)
	if !ok {
		return "", errors.New("original url not found")
	}
	s := userURL.(entity.UserURL).OriginalURL
	return s, nil
}

func (m *MapStorage) SetURL(shortURL string, originalURL string, userID string) (string, error) {

	userURL := entity.UserURL{
		UserID:      userID,
		OriginalURL: originalURL,
	}
	result, err := m.GetKey(userURL)
	switch {
	case errors.Is(err, internalerrors.ErrNotFound):
		{
			_, loaded := m.data.LoadOrStore(shortURL, userURL)
			if m.helper != nil {
				m.helper.WriteFile(lenSyncMap(m.data), shortURL, userURL)
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

func (m *MapStorage) SetURLBatch(u map[string]entity.UserURL) (map[string]entity.UserURL, error) {
	returned := make(map[string]entity.UserURL)
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
				returned[result] = u[s]
			}
		default:
			{
				return returned, err
			}
		}
	}
	return returned, possibleDoubleError
}

func (m *MapStorage) GetUserUrls(userID string) (any, error) {
	result := make([]entity.UserURLEntity, 0)
	m.data.Range(func(key, value interface{}) bool {
		if value.(entity.UserURL).UserID == userID {
			result = append(result, entity.UserURLEntity{
				ShortURL:    key.(string),
				OriginalURL: value.(entity.UserURL).OriginalURL})
			return true
		}
		return false
	})
	if len(result) == 0 {
		return nil, internalerrors.ErrNotFound
	}
	return result, nil
}

func (m *MapStorage) DeleteUserURLs(userID string) (deletedURLs chan string, err error) {
	deletedURLs = make(chan string)
	go func() {
		for key := range deletedURLs {
			m.data.Delete(key)
		}
		err := m.helper.RMFile(m.data)
		if err != nil {
			return
		}
	}()
	return deletedURLs, nil
}

func (m *MapStorage) GetKey(userURL entity.UserURL) (string, error) {
	var storedKey string
	ok := false
	m.data.Range(func(key, value interface{}) bool {
		if value.(entity.UserURL).OriginalURL == userURL.OriginalURL {
			storedKey = key.(string)
			ok = true
			return false
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
