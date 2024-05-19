package mw

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type GzipResponseWriter struct {
	Writer io.Writer
	http.ResponseWriter
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	write, err := w.Writer.Write(b)
	if err != nil {
		return 0, fmt.Errorf("error writing to gzip writer: %w", err)
	}
	return write, nil
}

func GzipMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle gzip request body
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer func(gz *gzip.Reader) {
				err := gz.Close()
				if err != nil {
					log.Println("Error closing gzip body")
				}
			}(gzipReader)
			r.Body = gzipReader
		}

		contentType := w.Header().Get("Content-Type")
		isCompressingContent := strings.HasPrefix(contentType, "application/json") ||
			strings.HasPrefix(contentType, "text/html")

		// Handle gzip response
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && isCompressingContent {
			gzipWriter := gzip.NewWriter(w)
			defer func(gzipWriter *gzip.Writer) {
				err := gzipWriter.Close()
				if err != nil {
					log.Println("Error closing gzip writer")
				}
			}(gzipWriter)

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			gzipResponseWriter := &GzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}
			next.ServeHTTP(gzipResponseWriter, r)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

/*import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

const (
	//https://reintech.io/blog/working-with-regular-expressions-in-go
	pattern = `https?://[^\s]+`
)

func New(h http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				rb, err := io.ReadAll(r.Body)
				defer r.Body.Close()
				rbstring := string(rb)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
				match, _ := regexp.MatchString(pattern, rbstring)
				if !match {
					http.Error(w, fmt.Sprintf("Bad URL, need pattern %s", pattern), http.StatusBadRequest)
					log.Printf("Bad URL, need pattern %s %s", pattern, err)
					return
				}
				//body теряется костыль, наверное https://www.reddit.com/r/golang/comments/mnht8z/getting_response_headers_and_body_in_middleware/
				r.Body = io.NopCloser(bytes.NewBuffer(rb))
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}*/
