package middlewares

import (
	"net/http"
	"stargazer/video-recording/config"
)

func MediaConvertAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Unauthorized: missing credentials", http.StatusUnauthorized)
			return
		}

		if username != config.App.MEDIA_CONVERT_BASIC_AUTH_USER || password != config.App.MEDIA_CONVERT_BASIC_AUTH_PASS {
			http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
