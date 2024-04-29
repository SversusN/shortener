package main

import (
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"regexp"
)

const (
	//https://reintech.io/blog/working-with-regular-expressions-in-go
	pattern = `https?://[^\s]+`
)

var uriCollection map[string]string

func main() {
	ParseFlags()
	//Эмуляция БД мапой
	uriCollection = make(map[string]string)
	//go client.GetClient() - клиент сохранен только локально

	/*mux := http.NewServeMux()
	mux.HandleFunc("/", router)*/
	log.Fatal(http.ListenAndServe(flagAdress, ChiRouter()))

}

func ChiRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlerPost)
		r.Get("/{shortKey}", handlerGet)
	})
	return r
}

// фнукция роутера, более не нужна
/*func router(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		handlerGet(res, req)
	} else if req.Method == http.MethodPost {
		handlerPost(res, req)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}*/

func handlerPost(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("Пришел %s \n ", req.Method)
	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Произошла ошибка при чтении ссылки %s", err)
		return
	}
	stringURL := string(originalURL)
	match, _ := regexp.MatchString(pattern, stringURL)
	if !match {
		http.Error(res, fmt.Sprintf("URL не соответствует формату %s", pattern), http.StatusBadRequest)
		log.Printf("URL не соответствует формату %s %s", pattern, err)
		return
	}
	if string(originalURL) != "" {
		shortKey := base64.StdEncoding.EncodeToString(originalURL)
		uriCollection[shortKey] = string(originalURL)
		shortenedURL := fmt.Sprint(flagBaseAdress, "/", shortKey)
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
	//shortKey := req.URL.Path[len("/"):]
	//Переделаем на URLParam в тестах пустой контекст
	shortKey := chi.URLParam(req, "shortKey")
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
// К сожалению плохо подходит для тестирования
// https://gosamples.dev/random-numbers/
/*func generateShortKey() string {
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}*/
