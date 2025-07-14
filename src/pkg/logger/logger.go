package logger

import (
	"log/slog"
	"os"

	"github.com/JaimeStill/persistent-context/pkg/config"
)

// Logger wraps the structured logger
type Logger struct {
	*slog.Logger
}

// New creates a new structured logger based on configuration
func New(cfg *config.LoggingConfig) *Logger {
	var handler slog.Handler
	
	// Create appropriate handler based on format
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: parseLogLevel(cfg.Level),
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: parseLogLevel(cfg.Level),
		})
	}
	
	logger := slog.New(handler)
	
	return &Logger{Logger: logger}
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithFields adds fields to the logger context
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{Logger: l.Logger.With("component", component)}
}

// WithRequestID adds a request ID field to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With("request_id", requestID)}
}

// Setup sets up the global logger
func Setup(cfg *config.LoggingConfig) *Logger {
	logger := New(cfg)
	
	// Set as default logger
	slog.SetDefault(logger.Logger)
	
	return logger
}