package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger to intercept log calls
type Logger struct {
	*slog.Logger
}

var Log *Logger

func sendToThirdParty(level, msg string, args ...any) {
	// TODO: Implement your third-party integration
}

// Info logs at Info level and sends to third-party tools
func (l *Logger) Info(msg string, args ...any) {
	sendToThirdParty("INFO", msg, args...)
	l.Logger.Info("✅ "+msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	sendToThirdParty("DEBUG", msg, args...)
	l.Logger.Debug("🔍"+msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	sendToThirdParty("ERROR", msg, args...)
	l.Logger.Error("❌ "+msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	sendToThirdParty("WARN", msg, args...)
	l.Logger.Warn("⚠️ "+msg, args...) // Warn
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	sendToThirdParty("INFO", msg, args...)
	l.Logger.InfoContext(ctx, "✅ "+msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	sendToThirdParty("DEBUG", msg, args...)
	l.Logger.DebugContext(ctx, "🔍"+msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	sendToThirdParty("ERROR", msg, args...)
	l.Logger.ErrorContext(ctx, "❌ "+msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	sendToThirdParty("WARN", "⚠️ "+msg, args...)
}

func init() {
	env := os.Getenv("ENVIRONMENT")

	if env == "" {
		env = "development"
	}

	var slogLogger *slog.Logger
	if env == "production" {
		// JSON format for production (better for AWS CloudWatch)
		slogLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		// Human-readable format for development
		slogLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	Log = &Logger{Logger: slogLogger}
}

// WithComponent returns a logger with a component field
func WithComponent(component string) *Logger {
	return &Logger{Logger: Log.With("component", component)}
}

// WithUser returns a logger with user context
func WithUser(username string) *Logger {
	return &Logger{Logger: Log.With("user", username)}
}

// WithError returns a logger with error context
func WithError(err error) *Logger {
	return &Logger{Logger: Log.With("error", err.Error())}
}

// WithRequest returns a logger with request context
func WithRequest(requestId string) *Logger {
	return &Logger{Logger: Log.With("request_id", requestId)}
}
