package middlewares

import (
	"net/http"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/api/helpers"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// List of allowed origins

		ErrMeta := helpers.GenerateSpan(r, w)
		defer ErrMeta.Trace.AddTraceToCtx(r)

		allowedOrigins := strings.Split(config.App.CORS_ALLOWED_ORIGINS, ",")

		// Get the origin of the incoming request
		origin := r.Header.Get("Origin")

		// Check if the request's Origin header is in the list of allowed origins
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Ensure that credentials are allowed
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// Allow specific methods
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// If it's an OPTIONS request, respond with no content (CORS preflight)
		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
