package middlewares

import (
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/comms"
)

func InterServiceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ErrMeta := helpers.GenerateSpan(r, w)
		defer ErrMeta.Trace.AddTraceToCtx(r)

		token, err := comms.ExtractToken(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			helpers.RespondInterServiceCallFailed(err, &ErrMeta)
			return
		}

		err = comms.ValidateForeignRequest(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			helpers.RespondInterServiceCallFailed(err, &ErrMeta)
			return
		}
		next.ServeHTTP(w, r)
	})
}
