package helpers

import (
	"net/http"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/internal/yad"
)

func RespondJsonDecodeErr(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetJsonDecodeError(err)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondJsonEncodeErr(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetJsonEncodeError(err)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondQueryParamsNotFound(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.QueryParamsNotFound, error_handler.QueryParamsNotFound, 400, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondIoErr(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.IOReadErr, error_handler.InternalError, http.StatusInternalServerError, false)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioRoomAlreadyExists(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioRoomAlreadyExist, error_handler.TwilioRoomAlreadyExist, http.StatusBadRequest, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioRoomAlreadyClosed(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioRoomAlreadyClosed, error_handler.TwilioRoomAlreadyClosed, http.StatusBadRequest, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioRoomRoomNotFound(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioRoomNotFound, error_handler.TwilioRoomNotFound, http.StatusNotFound, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioVideoNotFound(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioVideoNotFound, error_handler.TwilioVideoNotFound, http.StatusNotFound, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioRoomCreationError(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioRoomCloseError, error_handler.InternalError, http.StatusBadRequest, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondTwilioRoomCloseError(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.TwilioRoomAlreadyExist, err.Error(), http.StatusBadRequest, true)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondDbFailed(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.DBAddRetrieveFailed, err.Error(), http.StatusInternalServerError, false)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondValidationFailed(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, error_handler.ValidationFailed, err.Error(), http.StatusBadRequest, false)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func RespondInterServiceCallFailed(err error, ErrMeta *error_handler.ErrorResponseMeta) {
	var ErrResponse []byte
	ErrMeta.SetLogs(err, err.Error(), err.Error(), http.StatusUnauthorized, false)
	ErrResponse = error_handler.GenerateErrorResponse(ErrMeta)
	ErrMeta.Writer.Write(ErrResponse)
}

func SendErrRes(res []byte, w http.ResponseWriter) {
	w.Write(res)
}

func GenerateSpan(r *http.Request, w http.ResponseWriter) (ErrMeta error_handler.ErrorResponseMeta) {
	reqCtxData := r.Context().Value(yad.ReqCtxKey("trace")).(yad.ReqContext)
	trace := reqCtxData.Trace
	span := trace.CreateSpan(utils.GetCurrentFunctionName())
	trace.AddSpan(span)
	reqCtxData.Span = span
	ErrMeta.Init(w, reqCtxData)
	return ErrMeta
}
