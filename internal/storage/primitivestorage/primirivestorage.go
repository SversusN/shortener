// Пакет primitivestorage для работы с сохранением map в файл
package primitivestorage

import (
	"errors"
	"log"
	"sync"

	"github.com/SversusN/shortener/internal/internalerrors"
	"github.com/SversusN/shortener/internal/pkg/utils"
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
)

// MapStorage cnhernehf c потокобезопасной map и файлом
type MapStorage struct {
	data   *sync.Map
	helper *utils.FileHelper
}

// NewStorage хелпер межет придти nil, в этом случае сохранение в файл не работает
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

// GetURL реализация получения единичной ссылки
func (m *MapStorage) GetURL(id string) (string, error) {
	userURL, ok := m.data.Load(id)
	if !ok {
		return "", errors.New("original url not found")
	}
	s := userURL.(entity.UserURL).OriginalURL
	return s, nil
}

// SetURL реализация установки единичной ссылки
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

// SetURLBatch пакетное сохранение ссылок в файл
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

// GetUserUrls получение пользовательских ссылок по фильтру ИД пользователя
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

// DeleteUserURLs асинхронное удаление ссылок
func (m *MapStorage) DeleteUserURLs(userID string, group *sync.WaitGroup) (deletedURLs chan string, err error) {
	deletedURLs = make(chan string)
	group.Add(1)
	go func() {
		group.Done()
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

// GetKey получение ключа из sync.Map
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

// lenSyncMap размер map
func lenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}
