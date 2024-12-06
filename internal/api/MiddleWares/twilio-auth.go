package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/pkg/twilio"
)

func AuthenticateTwilio(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		params := make(map[string]string)
		signature := r.Header.Get("X-Twilio-Signature")
		for k, v := range r.Form {
			params[k] = v[0]
		}

		ok := twilio.ValidateRequest(signature, params)
		if ok {
			next.ServeHTTP(w, r)
			return
		}
		fmt.Println("das")
		data, _ := json.Marshal(schemas.ErrResponse{
			Error: "Request not originating from twilio",
		})
		w.WriteHeader(401)
		w.Write(data)
	})
}
