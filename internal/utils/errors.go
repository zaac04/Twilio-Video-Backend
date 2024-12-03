package utils

import (
	yad "stargazer/video-recording/internal/yad"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func CheckError(e error, Context string) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	if e != nil {
		yad.Logger.Error().Stack().Err(e).Msg(Context)
	}
}

func LogError(e error, reqId string, internalLog string, externalog string) {
	yad.Logger.Error().Stack().Err(e).Str("reqID", reqId).Str("ExternalLog", externalog).Msg(internalLog)
}
