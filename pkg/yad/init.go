package yad

import (
	"io"
	"log/slog"
	"os"
)

var Logger *slog.Logger
var ReqLogger *slog.Logger

type LineBreakWriter struct {
	writer io.Writer
}

func (w *LineBreakWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if err != nil {
		return n, err
	}
	_, err = w.writer.Write([]byte("\n"))
	return n, err
}

func InitializeLogger() {
	writer := &LineBreakWriter{writer: os.Stdout}
	jsonHandler := slog.NewJSONHandler(writer, nil)
	Logger = slog.New(jsonHandler)
	ReqLogger = slog.New(jsonHandler)
	Logger.Info("Initialized logger")
}
