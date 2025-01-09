package error_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/yad"
)

type ErrorResponseMeta struct {
	Trace       *yad.Trace
	Span        *yad.Span
	Writer      http.ResponseWriter
	Request     *http.Request
	Err         error
	StatusCode  int
	InternalLog string
	ExternalLog string
	Hint        bool `default:"false"`
	ReqId       string
}

func (e *ErrorResponseMeta) SetLogs(err error, InternalLog string, ExternalLog string, statusCode int, hint bool) {
	e.ExternalLog = ExternalLog
	e.InternalLog = InternalLog
	e.StatusCode = statusCode
	e.Err = err
}

func (e *ErrorResponseMeta) Init(w http.ResponseWriter, reqCtx yad.ReqContext) {
	e.Writer = w
	e.ReqId = reqCtx.ReqID
	e.Trace = reqCtx.Trace
	e.Span = reqCtx.Span
}

func (e *ErrorResponseMeta) SetHint(hint bool) {
	e.Hint = hint
}

func (e *ErrorResponseMeta) SetJsonEncodeError(err error) {
	e.Err = err
	e.ExternalLog = InternalError
	e.Hint = false
	e.InternalLog = ErrorEncodingJson
	e.StatusCode = http.StatusInternalServerError
	e.Span.AddEvent(ErrorEncodingJson, err.Error())
}

func (e *ErrorResponseMeta) SetValidationError(err error) {
	e.Err = err
	e.ExternalLog = ValidationFailed
	e.Hint = true
	e.InternalLog = err.Error()
	e.StatusCode = http.StatusBadRequest
	e.Span.AddEvent(ValidationFailed, err.Error())
}

func (e *ErrorResponseMeta) SetUnHandledInternalError(w http.ResponseWriter, reqCtx yad.ReqContext, err error) {
	e.Writer = w
	e.ReqId = reqCtx.ReqID
	e.ExternalLog = UnhandledInternalError
	e.Err = err
	e.Hint = true
	e.InternalLog = UnhandledInternalError
	e.StatusCode = http.StatusInternalServerError
	e.Span.AddEvent(UnhandledInternalError, err.Error())
}

func (e *ErrorResponseMeta) SetJsonDecodeError(err error) {
	e.Err = err
	e.ExternalLog = ErrorDecodingJson
	e.Hint = true
	e.InternalLog = ErrorDecodingJson
	e.StatusCode = http.StatusBadRequest
	e.Span.AddEvent(ErrorDecodingJson, err.Error())
}

func GenerateErrorResponse(meta *ErrorResponseMeta) (ErrResponse []byte) {
	New_Err := errors.New(meta.Err.Error())

	jsonRes := schemas.ErrResponse{
		Error: meta.ExternalLog,
	}
	if meta.Hint {
		jsonRes.Hint = New_Err.Error()
	}

	if meta.Err != nil {
		meta.Span.AddEvent(meta.InternalLog, meta.Err.Error())
	}

	ErrResponse, err := json.Marshal(jsonRes)
	if err != nil {
		meta.Span.AddEvent(ErrorEncodingJson, err)
	}
	meta.Writer.WriteHeader(meta.StatusCode)
	return ErrResponse
}
