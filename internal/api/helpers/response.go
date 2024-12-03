package helpers

import (
	"net/http"
)

func SendResponse(w http.ResponseWriter, res []byte) {
	w.Write(res)
}
