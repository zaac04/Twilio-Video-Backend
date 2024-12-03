package middlewares

import (
	"fmt"
	"net/http"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/structs"
)

func PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				reqCtxData := r.Context().Value(structs.ReqCtxKey("reqId")).(structs.ReqContext)
				var ErrorResponseMeta error_handler.ErrorResponseMeta
				ErrorResponseMeta.SetUnHandledInternalError(w, structs.ReqContext{ReqID: reqCtxData.ReqID}, fmt.Errorf("%v", err))
				res := error_handler.GenerateErrorResponse(&ErrorResponseMeta)
				w.Write(res)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
