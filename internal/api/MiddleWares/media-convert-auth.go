package middlewares

import (
	"errors"
	"net/http"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/api/helpers"
	error_handler "stargazer/video-recording/internal/error"
)

func MediaConvertAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ErrMeta := helpers.GenerateSpan(r, w)
		defer ErrMeta.Trace.AddTraceToCtx(r)

		username, password, ok := r.BasicAuth()
		if !ok {
			ErrMeta.Span.AddEvent(error_handler.MediaConvertAuthFailed, errors.New("unauthorized: missing credentials"))
			http.Error(w, "Unauthorized: missing credentials", http.StatusUnauthorized)
			return
		}

		if username != config.App.MEDIA_CONVERT_BASIC_AUTH_USER || password != config.App.MEDIA_CONVERT_BASIC_AUTH_PASS {
			ErrMeta.Span.AddEvent(error_handler.MediaConvertAuthFailed, errors.New("unauthorized: missing credentials"))
			http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
