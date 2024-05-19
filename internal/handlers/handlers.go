package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/storage"
)

type Handlers struct {
	cfg *config.Config
	s   storage.Storage
}

type JsonRequest struct {
	Url string `json:"url"`
}
type JsonResponse struct {
	Result string `json:"result"`
}

func NewHandlers(cfg *config.Config, s storage.Storage) *Handlers {
	return &Handlers{cfg, s}
}

func (h Handlers) HandlerPost(res http.ResponseWriter, req *http.Request) {
	//log.Printf("Request %s \n ", req.Method)
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

func (h Handlers) HandlerGet(res http.ResponseWriter, req *http.Request) {
	//log.Printf("Request %s \n ", req.Method)
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

func (h Handlers) HandlerJSONPost(res http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	var reqBody JsonRequest
	if err = json.Unmarshal(b, &reqBody); err != nil {
		log.Printf("Error parsing JSON request body: %s", err)
	}
	defer req.Body.Close()
	key := utils.GenerateShortKey()
	e := h.s.SetURL(key, reqBody.Url)
	if e != nil {
		log.Println("smth bad with data storage, mb double key ->", e)
	}
	shortURL := fmt.Sprint(h.cfg.FlagBaseAddress, "/", key)
	resBody, e := json.Marshal(JsonResponse{Result: shortURL})
	if e != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resBody)
}
