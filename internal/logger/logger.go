package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

type Logger struct {
	*slog.Logger
}

func New(level, format string) *Logger {
	var logLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With("request_id", requestID),
	}
}

func (l *Logger) LogRequest(method, path string, statusCode int, duration string, requestID string) {
	l.WithRequestID(requestID).Info("HTTP request",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration", duration,
	)
}

func (l *Logger) LogError(err error, msg string, requestID string) {
	l.WithRequestID(requestID).Error(msg, "error", err.Error())
}

type LogEntry struct {
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Fields    interface{} `json:"fields,omitempty"`
}

func (l *Logger) LogStructured(level, message string, fields map[string]interface{}, requestID string) {
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: fmt.Sprintf("%d", os.Getpid()),
		Fields:    fields,
	}

	if requestID != "" {
		entry.RequestID = requestID
	}

	jsonData, _ := json.Marshal(entry)
	fmt.Println(string(jsonData))
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value("logger").(*Logger); ok {
		return logger
	}
	return New("INFO", "json")
}

func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
