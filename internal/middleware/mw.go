package mw

import (
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
}
