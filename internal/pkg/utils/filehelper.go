package utils

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type Fields struct {
	OriginalURL string `json:"original_url"`
	ShortKey    string `json:"short_url"`
}

type FileHelper struct {
	file *os.File
}

func NewFileHelper(filename string) *FileHelper {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil
	}
	return &FileHelper{file: file}
}

func (fh FileHelper) WriteFile(originalURL string, shortKey string) {
	t := Fields{OriginalURL: originalURL, ShortKey: shortKey}
	jt, _ := json.Marshal(t)
	jt = append(jt, '\n')
	_, err := fh.file.Write(jt)
	if err != nil {
		return
	}
}

func (fh FileHelper) ReadFile() *sync.Map {

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
