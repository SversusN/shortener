package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/storage"
)

type Handlers struct {
	cfg *config.Config
	s   storage.Storage
}

func NewHandlers(cfg *config.Config, s storage.Storage) *Handlers {
	return &Handlers{cfg, s}
}

// HandlerPost PostURL godoc
//
//	@Summary		post URL
//	@Tags			shortener
//	@Description	post long URL string
//	@Accept			plain
//	@Produce		plain
//	@Param			long_url body		string	 		true    	"long URL"
//	@Success		200		{string}	string			"success"
//	@Success		201		{string}    string			"created"
//	@Success		500		{string}	string			"fail"
//	@Router			/ [post]
func (h Handlers) HandlerPost(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s \n ", req.Method)
	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URL %s", err)
		return
	}

	if len(originalURL) > 0 {
		//key := base64.StdEncoding.EncodeToString(originalURL) //слеши в base64 ломают url
		key := utils.GenerateShortKey()
		e := h.s.SetURL(key, string(originalURL))
		if e != nil {
			log.Println("smth bad with data storage, mb double key ->", e)
		}
		shortURL := fmt.Sprint(h.cfg.FlagBaseAddress, "/", key)

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

// HandlerGet GetShortURL godoc
//
//		@Summary		get URL
//		@Tags			shortener
//		@Description	get short URL string
//		@Produce		plain
//		@Param			key	path		string	    true 	"short URL"
//		@Success		200		{string}	string			"success"
//	    @Failure		404     {string}	string			"not found"
//		@Failure		500		{string}	string			"fail"
//		@Router			/{key} [get]
func (h Handlers) HandlerGet(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s \n ", req.Method)
	key := chi.URLParam(req, "shortKey")
	if key == "" {
		http.Error(res, "Shortened key is missing", http.StatusBadRequest)
		return
	}
	originalURL, err := h.s.GetURL(key)
	if err != nil {
		http.Error(res, "Shortened key not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
