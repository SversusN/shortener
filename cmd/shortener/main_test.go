package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_handlerGet(t *testing.T) {
	uriCollection = make(map[string]string)
	uriCollection["shortKey"] = "https://example.com"
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		shortKey    string
		originalURL string
		want        want
	}{
		{
			name:        "empty originUrl",
			shortKey:    "/",
			originalURL: "",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "no empty originUrl",
			shortKey:    "/shortKey",
			originalURL: "https://example.com",
			want: want{
				statusCode: 307,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.shortKey, nil)
			w := httptest.NewRecorder()
			handlerGet(w, r)
			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Код статуса не совпадает с ожидаемым")
			if tt.shortKey != "/" {
				assert.Equal(t, uriCollection["shortKey"], res.Header.Get("Location"), "Вернуласть не та ссылка или пусто")
			}
		})
	}
}

func Test_handlerPost(t *testing.T) {
	uriCollection = make(map[string]string)
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		originalUrl string
		want        want
	}{
		{
			name:        "empty originalUrl",
			originalUrl: "",
			want: want{
				contentType: "",
				statusCode:  400,
			},
		},
		{
			name:        "no empty request",
			originalUrl: "http://example.com",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.originalUrl))
			w := httptest.NewRecorder()
			handlerPost(w, r)
			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Код статуса не совпадает с ожидаемым")
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"), "Content-Type не совпадает с ожидаемым")
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err, "Неизвестная ошибка")
		})
	}
}
