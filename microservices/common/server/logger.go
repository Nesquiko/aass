package server

import (
	"log/slog"

	"github.com/go-chi/httplog/v2"
)

func SetupLogger(logLevel slog.Level) *httplog.Logger {
	logger := httplog.NewLogger("wac", httplog.Options{
		LogLevel: slog.Level(logLevel),
	})
	slog.SetDefault(logger.Logger)
	return logger
}
