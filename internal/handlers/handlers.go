// Handlers пакет для функционирования http-обработчиков
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/internalerrors"
	mw "github.com/SversusN/shortener/internal/middleware"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/dbstorage"
	"github.com/SversusN/shortener/internal/storage/storage"
)

// Handlers тип для внедрения зависимости
type Handlers struct {
	cfg         *config.Config
	s           storage.Storage
	waitGroup   *sync.WaitGroup
	trustSubnet *net.IPNet
}

// JSONRequest передача JSON Объекта в обработчик
type JSONRequest struct {
	URL string `json:"url"`
}

// JSONResponse JSON ответ
type JSONResponse struct {
	Result string `json:"result"`
}

// JSONBatchRequest пакет URL JSON формат запрос
type JSONBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// JSONBatchResponse пакет URL JSON формат ответ
type JSONBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortenedURL  string `json:"short_url"`
}

// JSONUserURLs ответ для пользовательских URL
type JSONUserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
type statsResponse struct {
	URLs  int `json:"urls"`  // количество сокращённых URL в сервисе
	Users int `json:"users"` // количество пользователей в сервисе
}

// NewHandlers инициализация объекта handlers
func NewHandlers(cfg *config.Config, s storage.Storage, waitGroup *sync.WaitGroup, ts *net.IPNet) *Handlers {
	return &Handlers{cfg, s, waitGroup, ts}
}

// HandlerPost получает оригинальный URL для сокращения в формате text\plain
func (h *Handlers) HandlerPost(res http.ResponseWriter, req *http.Request) {
	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URL %s", err)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	userID, err := getUserIDFromCtx(req)
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(originalURL) > 0 {
		var shortURL string
		key := utils.GenerateShortKey()
		var result string
		result, err = h.s.SetURL(key, string(originalURL), userID)

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

// HandlerGet получает сокращенный URL для перенаправления
func (h *Handlers) HandlerGet(res http.ResponseWriter, req *http.Request) {
	key := chi.URLParam(req, "shortKey")
	if key == "" {
		http.Error(res, "Shortened key is missing", http.StatusBadRequest)
		return
	}
	originalURL, err := h.s.GetURL(key)
	if errors.Is(err, internalerrors.ErrDeleted) {
		http.Error(res, err.Error(), http.StatusGone)
		return
	}
	if err != nil {
		http.Error(res, "Shortened key not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// HandlerJSONPost получает URL для сокращения в JSON формате
func (h *Handlers) HandlerJSONPost(res http.ResponseWriter, req *http.Request) {
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
	userID, err := getUserIDFromCtx(req)
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	key = utils.GenerateShortKey()
	var result string
	result, err = h.s.SetURL(key, reqBody.URL, userID)

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

// HandlerJSONPostBatch сохраняет пакет в JSON формате
func (h *Handlers) HandlerJSONPostBatch(res http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing URLs %s", err)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	defer req.Body.Close()
	var (
		reqBody  []JSONBatchRequest
		respBody []JSONBatchResponse
	)
	saveUrls := make(map[string]dbstorage.UserURL)
	if err = json.Unmarshal(b, &reqBody); err != nil {
		http.Error(res, "Bad JSON request...", http.StatusBadRequest)
	}
	userID, err := getUserIDFromCtx(req)
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, r := range reqBody {
		newKey := utils.GenerateShortKey()
		saveUrls[newKey] = dbstorage.UserURL{UserID: userID, OriginalURL: r.OriginalURL}
	}
	if len(saveUrls) > 0 {
		var rs JSONBatchResponse

		var mapResp map[string]dbstorage.UserURL

		mapResp, err = h.s.SetURLBatch(saveUrls)

		for s := range mapResp {
			i := indexOfURL(mapResp[s].OriginalURL, reqBody)
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

// HandlerDBPing проверяет возможность использования БД
func (h *Handlers) HandlerDBPing(res http.ResponseWriter, req *http.Request) {
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

// HandlerGetUserURLs получение пользователя из Cookie
func (h *Handlers) HandlerGetUserURLs(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie(mw.NameCookie)
	if err != nil {
		http.Error(w, "Bad Token, no token in cookie", http.StatusUnauthorized)
		return
	}
	userID, err := getUserIDFromCtx(r)
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		http.Error(w, "Bad userID, need Int data", http.StatusBadRequest)
		return
	}
	if userID == "" {
		http.Error(w, "No userID, bad token data", http.StatusUnauthorized)
		return
	}
	mapRest, err := h.s.GetUserUrls(userID)
	if errors.Is(err, internalerrors.ErrNotFound) {
		http.Error(w, "No URLs for user", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Bad userID", http.StatusUnauthorized)
		return
	}
	entities, ok := mapRest.([]dbstorage.UserURLEntity)
	if !ok {
		http.Error(w, "Bad userID, need Int data", http.StatusInternalServerError)
		return
	}
	if len(entities) == 0 {
		http.Error(w, "No URLs for user", http.StatusNotFound)
		return
	}
	var resBody []JSONUserURLs
	for _, o := range entities {
		resBody = append(resBody, JSONUserURLs{ShortURL: h.getFullURL(o.ShortURL), OriginalURL: o.OriginalURL})
	}
	resBodyJSON, err := json.Marshal(&resBody)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resBodyJSON)
}

// HandlerDeleteUserURLs - запускает процесс удаления URL
// Удаление происходит асинхронно
func (h *Handlers) HandlerDeleteUserURLs(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserIDFromCtx(r)
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		http.Error(w, "Bad userID, need Int data", http.StatusBadRequest)
	}
	if userID == "" {
		http.Error(w, "No userID, bad token data", http.StatusUnauthorized)
	}
	deleteURLs := make([]string, 0)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusInternalServerError)
	}
	err = json.Unmarshal(b, &deleteURLs)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusInternalServerError)
	}
	// Создается канал с наполнением URL для удаления
	deleteCh, err := h.s.DeleteUserURLs(userID, h.waitGroup)
	if err != nil {
		http.Error(w, "Bad userID", http.StatusBadRequest)
	}
	// Заполнение канала deleteCh для репозитория
	h.waitGroup.Add(1)
	go func() {
		defer h.waitGroup.Done()
		defer close(deleteCh)
		for _, key := range deleteURLs {
			deleteCh <- key
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) HandlerGetStats(w http.ResponseWriter, r *http.Request) {

	if h.trustSubnet == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !h.trustSubnet.Contains(ip) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	countUsers, countURLs, err := storage.Storage.GetStats(nil)
	if err != nil {
		statsResp := statsResponse{
			URLs:  0,
			Users: 0,
		}
		statsResp.Users = countUsers
		statsResp.URLs = countURLs

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(statsResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Can`t get from storage", http.StatusInternalServerError)
		return
	}
}

// indexOfURL получает индекс или возвращает -1
// -1 обозначает, что значение не было найдено
func indexOfURL(element string, data []JSONBatchRequest) int {
	for k, v := range data {
		if element == v.OriginalURL {
			return k
		}
	}
	return -1 //not found.
}

// getUserIDFromCtx попытка получить ИД пользователя из Cookie
func getUserIDFromCtx(r *http.Request) (string, error) {
	userID := r.Context().Value(mw.CtxUser)
	if userID == nil {
		return "", errors.New("user ID is missing")
	}
	userIDInt, ok := userID.(string)
	if !ok {
		return "", internalerrors.ErrUserTypeError
	} else {
		return userIDInt, nil
	}
}

// getFullURL - создает валидную полноценную ссылку из адреса и короткого ключа
func (h *Handlers) getFullURL(result string) string {
	return fmt.Sprint(h.cfg.FlagBaseAddress, "/", result)
}
