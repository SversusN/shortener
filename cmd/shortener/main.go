package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

const (
	charset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLength = 6
)

var uriCollection map[string]string

func main() {
	//Эмуляция БД мапой
	uriCollection = make(map[string]string)
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server listening on port 8080")
}

// роутер
func router(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		handlerGet(res, req)
	} else if req.Method == http.MethodPost {
		handlerPost(res, req)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func handlerPost(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("Пришел %s \n ", req.Method)
	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Произошла ошибка при чтении ссылки %s", err)
		return
	}
	if string(originalURL) != "" {
		shortKey := generateShortKey()
		uriCollection[shortKey] = string(originalURL)
		shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortenedURL))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}

}

func handlerGet(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Пришел GET")
	//возьмем слайс из URL после /
	shortKey := req.URL.Path[len("/"):]
	//не передали ключ
	if shortKey == "" {
		http.Error(res, "Shortened key is missing", http.StatusBadRequest)
		return
	}
	//Если не нашли то bad request
	originalURL, ok := uriCollection[shortKey]
	if !ok {
		http.Error(res, "Shortened key not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// Практический уникальный генератор
// https://gosamples.dev/random-numbers/
func generateShortKey() string {
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
