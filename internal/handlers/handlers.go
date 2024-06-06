package handlers

import (
	"context"
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

type JSONRequest struct {
	URL string `json:"url"`
}
type JSONResponse struct {
	Result string `json:"result"`
}
type JSONBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
type JSONBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortenedURL  string `json:"short_url"`
}

func NewHandlers(cfg *config.Config, s storage.Storage) *Handlers {
	return &Handlers{cfg, s}
}

func (h Handlers) getFullURL(result string) string {
	return fmt.Sprint(h.cfg.FlagBaseAddress, "/", result)
}

func (h Handlers) HandlerPost(res http.ResponseWriter, req *http.Request) {
	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URL %s", err)
		return
	}
	if len(originalURL) > 0 {
		var shortURL string
		result, err := h.s.GetKey(string(originalURL))
		if err != nil {
			key := utils.GenerateShortKey()
			e := h.s.SetURL(key, string(originalURL))
			if e != nil {
				log.Println("smth bad with data storage ->", e)
			}
			shortURL = h.getFullURL(key)
			res.WriteHeader(http.StatusCreated)
		} else {
			shortURL = h.getFullURL(result)
			res.WriteHeader(http.StatusConflict)
		}
		res.Header().Set("Content-Type", "text/plain")
		res.Write([]byte(shortURL))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (h Handlers) HandlerGet(res http.ResponseWriter, req *http.Request) {
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
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URLs %s", err)
		return
	}
	var (
		reqBody JSONRequest
		resBody JSONResponse
		key     string
	)
	if err = json.Unmarshal(b, &reqBody); err != nil {
		log.Printf("Error parsing JSON request body: %s", err)
		res.WriteHeader(http.StatusBadRequest)
	}
	defer req.Body.Close()
	key, err = h.s.GetKey(reqBody.URL)
	if err == nil {
		res.WriteHeader(http.StatusConflict)
		resBody.Result = h.getFullURL(key)
	} else {
		key = utils.GenerateShortKey()
		e := h.s.SetURL(key, reqBody.URL)
		if e != nil {
			log.Println("smth bad with data storage, mb double key ->", e)
		}
		resBody.Result = h.getFullURL(key)
		res.WriteHeader(http.StatusCreated)
	}
	resJSON, e := json.Marshal(&resBody)
	if e == nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resJSON)
}

func (h Handlers) HandlerJSONPostBatch(res http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URLs %s", err)
		return
	}
	ctx := context.Background()
	defer req.Body.Close()
	var reqBody []JSONBatchRequest
	var respBody []JSONBatchResponse
	saveUrls := make(map[string]string)
	if err = json.Unmarshal(b, &reqBody); err != nil {
		http.Error(res, "Bad JSON request...", http.StatusBadRequest)
	}
	for _, r := range reqBody {
		var rs JSONBatchResponse
		key, err := h.s.GetKey(r.OriginalURL)
		if err == nil {
			rs.ShortenedURL = h.getFullURL(key)
			res.WriteHeader(http.StatusConflict)
		} else {
			newKey := utils.GenerateShortKey()
			saveUrls[newKey] = r.OriginalURL
			rs.ShortenedURL = h.getFullURL(newKey)
			res.WriteHeader(http.StatusCreated)
		}
		rs.CorrelationID = r.CorrelationID
		respBody = append(respBody, rs)
	}
	if len(saveUrls) > 0 {
		err := h.s.SetURLBatch(ctx, saveUrls)
		//Если 2 одинаковых ссылки в пакете - транзакция упадет
		if err != nil {
			res.WriteHeader(http.StatusConflict)
			http.Error(res, err.Error(), http.StatusConflict)
			return
		}
	}
	resBody, err := json.Marshal(&respBody)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
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
