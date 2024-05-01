package main

import (
	"github.com/SversusN/shortener/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	//Подсказали в конференции =D
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
	ts := httptest.NewServer(a.CreateRouter(*a.Handlers))
	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		body         string
		path         string
		expectedCode int
		Location     string
	}{
		{
			name:         "Good Post request, 201 waiting",
			method:       http.MethodPost,
			body:         "http://example.com",
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
			path:         "/aHR0cDovL2V4YW1wbGUuY29t",
			expectedCode: http.StatusTemporaryRedirect,
			Location:     "http://example.com",
		},
		{
			name:         "Method on is not allowed",
			method:       http.MethodPost,
			path:         "/someBadKey",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "No short key URL",
			method:       http.MethodGet,
			path:         "/",
			expectedCode: http.StatusMethodNotAllowed,
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
		})
	}
}
