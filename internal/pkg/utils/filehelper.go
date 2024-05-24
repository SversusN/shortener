package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/SversusN/shortener/config"
)

type Fields struct {
	UUID        int    `json:"uuid"`
	OriginalURL string `json:"original_url"`
	ShortKey    string `json:"short_url"`
}

type FileHelper struct {
	file *os.File
	c    *config.Config
	mu   sync.Mutex //https://echorand.me/posts/go-file-mutex/#notes
}

// возвращаем хелпер или ошибку, чтобы выключить сохранение в файл

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
func (fh FileHelper) WriteFile(uuid int, originalURL string, shortKey string) {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	t := Fields{UUID: uuid, OriginalURL: originalURL, ShortKey: shortKey}
	jt, _ := json.Marshal(t)
	jt = append(jt, '\n')
	_, err := fh.file.Write(jt)
	if err != nil {
		return
	}
}

func (fh FileHelper) ReadFile() *sync.Map {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	tempMap := sync.Map{}
	if fh.file != nil {
		var fields Fields
		scanner := bufio.NewScanner(fh.file)
		for scanner.Scan() {
			err := json.Unmarshal(scanner.Bytes(), &fields)
			if err != nil {
				return nil
			}
			tempMap.LoadOrStore(fields.ShortKey, fields.OriginalURL)
		}
	}
	return &tempMap
}
