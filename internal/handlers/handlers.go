package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
)

type Handlers struct {
	cfg *config.Config
	s   storage.Storage
}

type JSONRequest struct {
	URL string `json:"url"`
}
type JSONResponse struct {
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
		var shortURL string
		result, err := h.s.GetKey(string(originalURL))
		//key := base64.StdEncoding.EncodeToString(originalURL) //слеши в base64 ломают url
		if err != nil {
			key := utils.GenerateShortKey()
			e := h.s.SetURL(key, string(originalURL))
			if e != nil {
				log.Println("smth bad with data storage, mb double key ->", e)
			}
			shortURL = fmt.Sprint(h.cfg.FlagBaseAddress, "/", key)
		} else {
			shortURL = fmt.Sprint(h.cfg.FlagBaseAddress, "/", result)
		}
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
	var reqBody JSONRequest
	if err = json.Unmarshal(b, &reqBody); err != nil {
		//не json
		log.Printf("Error parsing JSON request body: %s", err)
		res.WriteHeader(http.StatusBadRequest)
	}
	defer req.Body.Close()
	result, err := h.s.GetKey(reqBody.URL)
	var shortURL string
	if err != nil {
		key := utils.GenerateShortKey()
		e := h.s.SetURL(key, reqBody.URL)
		if e != nil {
			log.Println("smth bad with data storage, mb double key ->", e)
		}
		shortURL = fmt.Sprint(h.cfg.FlagBaseAddress, "/", key)
	}
	shortURL = fmt.Sprint(h.cfg.FlagBaseAddress, "/", result)
	resBody, e := json.Marshal(JSONResponse{Result: shortURL})
	if e != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resBody)
}

func (h Handlers) HandlerDBPing(res http.ResponseWriter, req *http.Request) {
	pinger, ok := h.s.(storage.Pinger)
	if !ok {
		http.Error(res, "No DB to ping , sorry...", http.StatusBadRequest)
		return
	}
	result := pinger.Ping()
	res.Header().Set("Content-Type", "text/plain")
	if result == nil {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK ping"))
	} else {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("BAD ping"))
	}
}
