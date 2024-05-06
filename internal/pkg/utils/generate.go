package utils

import (
	"crypto/rand"
	"fmt"
	"log"
)

func GenerateShortKey() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("smth bad with generate %v", err)
	}
	return fmt.Sprintf("%X", b[0:])
}
