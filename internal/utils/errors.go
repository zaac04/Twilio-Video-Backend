package utils

import (
	yad "stargazer/video-recording/pkg/yad"
)

func LogError(err error, reqID string, internalLog string, externalLog string) {
	yad.Logger.Error(
		internalLog,
		"error", err,
		"reqID", reqID,
		"externalLog", externalLog,
	)
}
