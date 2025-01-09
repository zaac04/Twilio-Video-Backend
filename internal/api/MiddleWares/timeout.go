package middlewares

import (
	"context"
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
	"time"
)

func Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ErrMeta := helpers.GenerateSpan(r, w)
		defer ErrMeta.Trace.AddTraceToCtx(r)

		timeOut, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		r = r.WithContext(timeOut)

		next.ServeHTTP(w, r)
	})
}
