package yad

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger
var ReqLogger *slog.Logger

func InitializeLogger() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	Logger = slog.New(jsonHandler)
	ReqLogger = slog.New(jsonHandler)
	Logger.Info("Initialized logger")
}
