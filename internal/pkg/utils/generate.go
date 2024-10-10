// Пакет вспомогательных функций
package utils

import (
	"crypto/rand"
	"fmt"
	"log"
)

// GenerateShortKey формирует ключ короткой ссылки
func GenerateShortKey() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("smth bad with generate %v", err)
	}
	return fmt.Sprintf("%X", b[0:])
}

// GetFullURL getFullURL - создает валидную полноценную ссылку из адреса и короткого ключа
func GetFullURL(baseURL string, result string) string {
	return fmt.Sprint(baseURL, "/", result)
}
