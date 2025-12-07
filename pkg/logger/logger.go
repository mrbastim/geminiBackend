package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

var L *slog.Logger

func init() {
	// Настройка логгера по умолчанию, если Init() не вызывается
	L = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func Init(logLevel string, logFile string) {
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler *slog.TextHandler
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		multiWriter := io.MultiWriter(os.Stdout, file)
		handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}
	L = slog.New(handler)
}
