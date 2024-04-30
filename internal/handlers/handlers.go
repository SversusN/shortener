package handlers

import (
	"encoding/base64"
	"fmt"
	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/storage/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"regexp"
)

type Handlers struct {
	config       *config.Config
	localstorage storage.Storage
}

func NewHandlers(config *config.Config, localstorage storage.Storage) *Handlers {
	return &Handlers{config, localstorage}
}

const (
	//https://reintech.io/blog/working-with-regular-expressions-in-go
	pattern = `https?://[^\s]+`
)

func (h Handlers) HandlerPost(res http.ResponseWriter, req *http.Request) {
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
		//uriCollection[shortKey] = string(originalURL)
		err := h.localstorage.SetURL(shortKey, string(originalURL))
		if err != nil {
			log.Println("smth bad with datastorage, $v", err)
		}
		shortenedURL := fmt.Sprint(h.config.FlagBaseAddress, "/", shortKey)

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortenedURL))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}

}

func (h Handlers) HandlerGet(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Пришел GET")
	shortKey := chi.URLParam(req, "shortKey")
	if shortKey == "" {
		http.Error(res, "Shortened key is missing", http.StatusBadRequest)
		return
	}

	originalURL, err := h.localstorage.GetURL(shortKey)
	if err != nil {
		http.Error(res, "Shortened key not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
