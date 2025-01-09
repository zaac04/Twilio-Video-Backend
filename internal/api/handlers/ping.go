package handlers

import (
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	helpers.GenerateSpan(r, w)
	w.WriteHeader(http.StatusOK)
	helpers.SendResponse(w, []byte("pong"))
}
