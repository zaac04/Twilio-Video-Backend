package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"stargazer/video-recording/internal/structs"
	"stargazer/video-recording/internal/utils"
	yad "stargazer/video-recording/internal/yad"
	"time"

	"github.com/google/uuid"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqId := uuid.New().String()
		ctx := context.WithValue(r.Context(), structs.ReqCtxKey("reqId"), structs.ReqContext{ReqID: reqId})
		r = r.WithContext(ctx)
		recorder := httptest.NewRecorder()
		ip := utils.GetClientIP(r)
		defer func() {
			yad.ReqLogger.Info().Str("method", r.Method).Str("ip", ip).Str("reqId", reqId).
				Str("reqUrl", r.URL.RequestURI()).Dur("elapsed", time.Since(start)).
				Int("StatusCode", recorder.Code).
				Str("ResponseBody", recorder.Body.String()).Msg("IN")
		}()
		next.ServeHTTP(w, r)
	})
}
