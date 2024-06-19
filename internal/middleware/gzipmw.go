package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type GzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
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
			defer func(gzipReader *gzip.Reader) {
				err := gzipReader.Close()
				if err != nil {
					http.Error(w, "Invalid close reader", http.StatusInternalServerError)
					return
				}
			}(gzipReader)
			r.Body = gzipReader
		}

		contentType := w.Header().Get("Content-Type")
		var isText = strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") // Handle gzip response
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && isText {
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
