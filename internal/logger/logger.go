package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

type Logger struct {
	*slog.Logger
}

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
)

type LogFields struct {
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Status    int    `json:"status,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Error     string `json:"error,omitempty"`
}

func New() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{logger}
}

func (l *Logger) WithFields(fields LogFields) *Logger {
	attrs := []slog.Attr{
		slog.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	}

	if fields.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", fields.RequestID))
	}
	if fields.UserID != "" {
		attrs = append(attrs, slog.String("user_id", fields.UserID))
	}
	if fields.Method != "" {
		attrs = append(attrs, slog.String("method", fields.Method))
	}
	if fields.Path != "" {
		attrs = append(attrs, slog.String("path", fields.Path))
	}
	if fields.Status != 0 {
		attrs = append(attrs, slog.Int("status", fields.Status))
	}
	if fields.Duration != "" {
		attrs = append(attrs, slog.String("duration", fields.Duration))
	}
	if fields.Error != "" {
		attrs = append(attrs, slog.String("error", fields.Error))
	}

	return &Logger{l.WithAttrs(attrs)}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := LogFields{}

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields.RequestID = requestID.(string)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		fields.UserID = userID.(string)
	}

	return l.WithFields(fields)
}

func (l *Logger) LogRequest(method, path string, status int, duration time.Duration, err error) {
	fields := LogFields{
		Method:   method,
		Path:     path,
		Status:   status,
		Duration: duration.String(),
	}

	if err != nil {
		fields.Error = err.Error()
	}

	l.WithFields(fields).Info("HTTP request")
}

func (l *Logger) LogError(err error, msg string) {
	l.WithFields(LogFields{Error: err.Error()}).Error(msg)
}

func (l *Logger) LogInfo(msg string, args ...interface{}) {
	l.Info(fmt.Sprintf(msg, args...))
}

func (l *Logger) LogDebug(msg string, args ...interface{}) {
	l.Debug(fmt.Sprintf(msg, args...))
}

func (l *Logger) LogWarn(msg string, args ...interface{}) {
	l.Warn(fmt.Sprintf(msg, args...))
}

// Context helpers
func WithRequestID(ctx context.Context) context.Context {
	requestID := uuid.New().String()
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		return requestID.(string)
	}
	return ""
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	if userID := ctx.Value(UserIDKey); userID != nil {
		return userID.(string)
	}
	return ""
}

// Metrics helper for structured logging
type Metrics struct {
	RequestsTotal    int64
	RequestDuration  time.Duration
	ErrorsTotal      int64
	ActiveConnections int64
}

func (l *Logger) LogMetrics(metrics Metrics) {
	fields := LogFields{
		"requests_total":     metrics.RequestsTotal,
		"request_duration":   metrics.RequestDuration.String(),
		"errors_total":       metrics.ErrorsTotal,
		"active_connections": metrics.ActiveConnections,
	}

	l.WithFields(fields).Info("application_metrics")
}