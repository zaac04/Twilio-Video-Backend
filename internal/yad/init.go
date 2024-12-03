package yad

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger
var ReqLogger zerolog.Logger

func InitializeLogger() {
	multiWriter := zerolog.MultiLevelWriter(os.Stdout)
	multiWriter_req := zerolog.MultiLevelWriter(os.Stdout)
	Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
	ReqLogger = zerolog.New(multiWriter_req).With().Timestamp().Logger()
	Logger.Info().Msg("Intialized logger")
}
