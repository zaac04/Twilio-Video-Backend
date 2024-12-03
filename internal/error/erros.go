package error_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/structs"
	"stargazer/video-recording/internal/utils"

	"strings"
)

type ErrorResponseMeta struct {
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

func (e *ErrorResponseMeta) Init(w http.ResponseWriter, reqCtx structs.ReqContext) {
	e.Writer = w
	e.ReqId = reqCtx.ReqID
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
}
func (e *ErrorResponseMeta) SetValidationError(err error) {
	e.Err = err
	e.ExternalLog = ValidationFailed
	e.Hint = true
	e.InternalLog = err.Error()
	e.StatusCode = http.StatusBadRequest
}

func (e *ErrorResponseMeta) SetFileDecodeError() {
	e.Err = errors.New(strings.ToLower(FileDecodeFailed))
	e.ExternalLog = FileDecodeFailed
	e.Hint = false
	e.InternalLog = FileDecodeFailed
	e.StatusCode = http.StatusBadRequest
}

func (e *ErrorResponseMeta) SetUnHandledInternalError(w http.ResponseWriter, reqCtx structs.ReqContext, err error) {
	e.Writer = w
	e.ReqId = reqCtx.ReqID
	e.ExternalLog = UnhandledInternalError
	e.Err = err
	e.Hint = true
	e.InternalLog = UnhandledInternalError
	e.StatusCode = http.StatusInternalServerError
}

func (e *ErrorResponseMeta) SetJsonDecodeError(err error) {
	e.Err = err
	e.ExternalLog = ErrorDecodingJson
	e.Hint = true
	e.InternalLog = ErrorDecodingJson
	e.StatusCode = http.StatusBadRequest
}

func GenerateErrorResponse(meta *ErrorResponseMeta) (ErrResponse []byte) {
	New_Err := errors.New(meta.Err.Error())
	jsonRes := schemas.ErrResponse{
		Error: meta.ExternalLog,
	}
	if meta.Hint {
		jsonRes.Hint = New_Err.Error()
	}
	utils.LogError(meta.Err, meta.ReqId, meta.InternalLog, meta.ExternalLog)
	ErrResponse, err := json.Marshal(jsonRes)
	utils.LogError(err, meta.ReqId, ErrorEncodingJson, meta.ExternalLog)
	meta.Writer.WriteHeader(meta.StatusCode)
	return ErrResponse
}
