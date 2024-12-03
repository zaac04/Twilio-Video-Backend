package handlers

import (
	"fmt"
	"net/http"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("pong"))
}

func MediaConvertCallback(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Unauthorized: missing credentials", http.StatusUnauthorized)
		return
	}

	if username != "abc" || password != "123" {
		http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
		return
	}

	fmt.Println(username, password, ok)

	var callback schemas.MediaConvertEvent
	err := utils.UnmarshalReqBody(r.Body, &callback)

	fmt.Println(err, callback)
}
