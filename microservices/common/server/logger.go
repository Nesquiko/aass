package server

import (
	"log/slog"

	"github.com/go-chi/httplog/v2"
)

func SetupLogger(name string, logLevel slog.Level) *httplog.Logger {
	logger := httplog.NewLogger(name, httplog.Options{
		LogLevel: slog.Level(logLevel),
	})
	slog.SetDefault(logger.Logger)
	return logger
}
