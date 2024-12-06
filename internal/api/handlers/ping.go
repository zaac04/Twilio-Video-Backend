package handlers

import (
	"net/http"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("pong"))
}
