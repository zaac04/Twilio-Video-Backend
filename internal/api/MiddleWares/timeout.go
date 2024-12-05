package middlewares

import (
	"context"
	"net/http"
	"time"
)

func Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeOut, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		r = r.WithContext(timeOut)
		next.ServeHTTP(w, r)
	})
}
