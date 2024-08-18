// Пакет подключения хранения в файле
package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/SversusN/shortener/config"
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
)

// Fields Поля объекта хранения URL
type Fields struct {
	UUID     int            `json:"uuid"`
	UserURL  entity.UserURL `json:"user_url"`
	ShortKey string         `json:"short_url"`
}

// FileHelper структура для работы с файлом
type FileHelper struct {
	file *os.File
	c    *config.Config
}

// NewFileHelper возвращаем хелпер или ошибку, чтобы выключить сохранение в файл
func NewFileHelper(filename string) (*FileHelper, error) {
	if filename == "" {
		return nil, errors.New("filename is empty, no store tempdb")
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileHelper{file: file}, nil
}

// WriteFile запись map в файл для сохранения после рестарта.
func (fh FileHelper) WriteFile(uuid int, shortURL string, userURL entity.UserURL) {
	t := Fields{UUID: uuid, ShortKey: shortURL, UserURL: userURL}
	jt, _ := json.Marshal(t)
	jt = append(jt, '\n')
	_, err := fh.file.Write(jt)
	if err != nil {
		return
	}
}

// RMFile Перезаписываем файл после того как отработает удаление
func (fh FileHelper) RMFile(data *sync.Map) error {
	err := os.Truncate(fh.file.Name(), 0)
	if err != nil {
		return errors.New("cannot remove file")
	}
	tmpMap := make(map[string]entity.UserURL)
	//https://stackoverflow.com/questions/46390409/how-to-decode-json-strings-to-sync-map-instead-of-normal-map-in-go1-9
	data.Range(func(k, v interface{}) bool {
		if k != nil {
			tmpMap[k.(string)] = v.(entity.UserURL)
		}
		return true
	})
	i := 1
	for k, v := range tmpMap {
		fh.WriteFile(i, k, v)
		i++
	}
	return nil
}

// ReadFile чтение файла для формирования sync.Map
func (fh FileHelper) ReadFile() *sync.Map {
	tempMap := sync.Map{}
	if fh.file != nil {
		var fields Fields
		scanner := bufio.NewScanner(fh.file)
		for scanner.Scan() {
			err := json.Unmarshal(scanner.Bytes(), &fields)
			if err != nil {
				//при пустом файле data - nil
				return &tempMap
			}
			tempMap.LoadOrStore(fields.ShortKey, fields.UserURL)
		}
	}
	return &tempMap
}
