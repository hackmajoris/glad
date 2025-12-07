package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	if env == "production" {
		// JSON format for production (better for AWS CloudWatch)
		Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		// Human-readable format for development
		Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
}

// WithComponent returns a logger with a component field
func WithComponent(component string) *slog.Logger {
	return Log.With("component", component)
}

// WithUser returns a logger with user context
func WithUser(username string) *slog.Logger {
	return Log.With("user", username)
}

// WithError returns a logger with error context
func WithError(err error) *slog.Logger {
	return Log.With("error", err.Error())
}

// WithRequest returns a logger with request context
func WithRequest(requestId string) *slog.Logger {
	return Log.With("request_id", requestId)
}
