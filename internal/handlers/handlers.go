package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SversusN/shortener/internal/internalerrors"
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
	res.Header().Set("Content-Type", "text/plain")
	if len(originalURL) > 0 {
		var shortURL string
		key := utils.GenerateShortKey()
		result, err := h.s.SetURL(key, string(originalURL))
		switch {
		case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
			res.WriteHeader(http.StatusConflict)
		case err != nil:
			res.WriteHeader(http.StatusInternalServerError)
		default:
			res.WriteHeader(http.StatusCreated)
		}
		shortURL = h.getFullURL(result)
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
	}
	res.Header().Set("Content-Type", "application/json")
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
	key = utils.GenerateShortKey()
	result, err := h.s.SetURL(key, reqBody.URL)
	switch {
	case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
		res.WriteHeader(http.StatusConflict)
	case err != nil:
		res.WriteHeader(http.StatusInternalServerError)
	default:
		res.WriteHeader(http.StatusCreated)
	}
	resBody.Result = h.getFullURL(result)

	resJSON, err := json.Marshal(&resBody)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Write(resJSON)
}

func (h Handlers) HandlerJSONPostBatch(res http.ResponseWriter, req *http.Request) {

	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URLs %s", err)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	defer req.Body.Close()
	var reqBody []JSONBatchRequest
	var respBody []JSONBatchResponse
	saveUrls := make(map[string]string)
	if err = json.Unmarshal(b, &reqBody); err != nil {
		http.Error(res, "Bad JSON request...", http.StatusBadRequest)
	}
	for _, r := range reqBody {
		newKey := utils.GenerateShortKey()
		saveUrls[newKey] = r.OriginalURL
	}
	if len(saveUrls) > 0 {
		var rs JSONBatchResponse
		mapResp, err := h.s.SetURLBatch(saveUrls)
		for s := range mapResp {
			i := indexOfURL(mapResp[s], reqBody)
			rs.ShortenedURL = h.getFullURL(s)
			rs.CorrelationID = reqBody[i].CorrelationID
			respBody = append(respBody, rs)
		}
		switch {
		case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
			res.WriteHeader(http.StatusConflict)
		case err != nil:
			res.WriteHeader(http.StatusInternalServerError)
		default:
			res.WriteHeader(http.StatusCreated)
		}
	}
	resBody, err := json.Marshal(&respBody)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
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

func indexOfURL(element string, data []JSONBatchRequest) int {
	for k, v := range data {
		if element == v.OriginalURL {
			return k
		}
	}
	return -1 //not found.
}
