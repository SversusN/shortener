package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SversusN/shortener/internal/app"
	"github.com/SversusN/shortener/internal/handlers"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			recover()
			log.Fatal("Error closing body")
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	a := app.New()
	//хенлеры проверяем не портим БД
	a.Storage = nil
	a.Storage = primitivestorage.NewStorage(nil, errors.New("dont need file"))
	//Для хендлеров тоже мап
	a.Handlers = handlers.NewHandlers(a.Config, a.Storage)
	a.Storage.SetURL("sk", "http://example.com", uuid.NewString())
	ts := httptest.NewServer(a.CreateRouter(*a.Handlers))

	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		body         string
		path         string
		expectedCode int
		Location     string
		contentType  string
	}{
		{
			name:         "Good Post request, 201 waiting",
			method:       http.MethodPost,
			body:         "http://example_example2.com",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Status 404 if no URL",
			method:       http.MethodGet,
			path:         "/shortBadKey",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Good request with shortKey",
			method:       http.MethodGet,
			path:         "/sk",
			expectedCode: http.StatusTemporaryRedirect,
			Location:     "http://example.com",
		},
		{
			name:         "Method on is not allowed",
			method:       http.MethodPost,
			path:         "/someBadKey",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "No short key URL",
			method:       http.MethodGet,
			path:         "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Json serialize handler test",
			method:       http.MethodPost,
			body:         "{\"url\":\"http://example_example_example.com\"}",
			path:         "/api/shorten",
			expectedCode: http.StatusCreated,
			contentType:  "application/json",
		},
		{
			name:         "Json batch handler test",
			method:       http.MethodPost,
			body:         "[\n{\n\"correlation_id\": \"1\",\n        \"original_url\": \"http://example33.com\"\n    }\n\n] ",
			path:         "/api/shorten/batch",
			expectedCode: http.StatusCreated,
			contentType:  "application/json",
		},
		{
			name:         "Get user URL Test. Bad NO auth",
			method:       http.MethodGet,
			path:         "/api/user/urls",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			resp, _ := testRequest(t, ts, tc.method, tc.path, tc.body)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Response code is not correct")

			if tc.Location != "" {
				assert.Equal(t, tc.Location, resp.Header.Get("Location"))
			}
			if tc.name == "Json serialize handler test" {
				assert.Equal(t, tc.contentType, resp.Header.Get("Content-Type"))
			}
			if tc.name == "Json batch handler test" {
				assert.Equal(t, tc.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
	//на всякий обнуляем конвеер
	a = nil
}
