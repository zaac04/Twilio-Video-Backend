package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"stargazer/video-recording/internal/utils"
	yad "stargazer/video-recording/pkg/yad"
	"strconv"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace := yad.StartNewTrace()
		ctx := context.WithValue(r.Context(), yad.ReqCtxKey("trace"), yad.ReqContext{ReqID: trace.TraceID, Trace: trace})
		r = r.WithContext(ctx)

		recorder := httptest.NewRecorder()
		ip := utils.GetClientIP(r)

		defer func() {

			trace = r.Context().Value(yad.ReqCtxKey("trace")).(yad.ReqContext).Trace

			trace = &yad.Trace{
				TraceID:   trace.TraceID,
				StartTime: trace.StartTime,
				EndTime:   time.Now(),
				Duration:  time.Since(trace.StartTime),
				Status: yad.Status{
					Code:    strconv.Itoa(recorder.Code),
					Message: recorder.Body.String(),
				},
				Attributes: map[string]string{
					"reqUrl": r.URL.RequestURI(),
					"ip":     ip,
					"method": r.Method,
				},
				Spans: trace.Spans,
			}
			data, _ := json.MarshalIndent(trace, "", "")
			yad.ReqLogger.Info("IN", "Trace", json.RawMessage(data))
		}()

		next.ServeHTTP(w, r)
	})
}
