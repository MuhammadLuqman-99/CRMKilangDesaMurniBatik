// Package logger provides structured logging utilities for the CRM application.
// It wraps the zerolog library to provide a consistent logging interface
// with support for contextual fields, log levels, and output formatting.
package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger to provide additional functionality.
type Logger struct {
	zl zerolog.Logger
}

// Config holds the logger configuration.
type Config struct {
	Level      string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format     string `yaml:"format" env:"LOG_FORMAT" env-default:"json"` // json or console
	TimeFormat string `yaml:"time_format" env:"LOG_TIME_FORMAT" env-default:"2006-01-02T15:04:05.000Z07:00"`
	Caller     bool   `yaml:"caller" env:"LOG_CALLER" env-default:"false"`
	Output     io.Writer
}

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	loggerKey contextKey = "logger"
)

// Global logger instance
var globalLogger *Logger

// init initializes the global logger with default settings.
func init() {
	globalLogger = New(Config{
		Level:      "info",
		Format:     "json",
		TimeFormat: time.RFC3339Nano,
		Caller:     false,
	})
}

// New creates a new Logger with the given configuration.
func New(cfg Config) *Logger {
	// Set time format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set output
	var output io.Writer
	if cfg.Output != nil {
		output = cfg.Output
	} else {
		output = os.Stdout
	}

	// Configure output format
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: cfg.TimeFormat,
		}
	}

	// Create base logger
	zl := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	// Add caller if enabled
	if cfg.Caller {
		zl = zl.With().Caller().Logger()
	}

	return &Logger{zl: zl}
}

// SetGlobal sets the global logger instance.
func SetGlobal(l *Logger) {
	globalLogger = l
}

// Global returns the global logger instance.
func Global() *Logger {
	return globalLogger
}

// WithContext returns a new context with the logger attached.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the logger from the context, or the global logger if not found.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	return globalLogger
}

// With returns a new logger with the given fields.
func (l *Logger) With() *LoggerContext {
	return &LoggerContext{zc: l.zl.With()}
}

// LoggerContext is a builder for adding fields to a logger.
type LoggerContext struct {
	zc zerolog.Context
}

// Str adds a string field.
func (lc *LoggerContext) Str(key, val string) *LoggerContext {
	lc.zc = lc.zc.Str(key, val)
	return lc
}

// Int adds an integer field.
func (lc *LoggerContext) Int(key string, val int) *LoggerContext {
	lc.zc = lc.zc.Int(key, val)
	return lc
}

// Int64 adds an int64 field.
func (lc *LoggerContext) Int64(key string, val int64) *LoggerContext {
	lc.zc = lc.zc.Int64(key, val)
	return lc
}

// Float64 adds a float64 field.
func (lc *LoggerContext) Float64(key string, val float64) *LoggerContext {
	lc.zc = lc.zc.Float64(key, val)
	return lc
}

// Bool adds a boolean field.
func (lc *LoggerContext) Bool(key string, val bool) *LoggerContext {
	lc.zc = lc.zc.Bool(key, val)
	return lc
}

// Time adds a time field.
func (lc *LoggerContext) Time(key string, val time.Time) *LoggerContext {
	lc.zc = lc.zc.Time(key, val)
	return lc
}

// Dur adds a duration field.
func (lc *LoggerContext) Dur(key string, val time.Duration) *LoggerContext {
	lc.zc = lc.zc.Dur(key, val)
	return lc
}

// Err adds an error field.
func (lc *LoggerContext) Err(err error) *LoggerContext {
	lc.zc = lc.zc.Err(err)
	return lc
}

// Interface adds an interface field.
func (lc *LoggerContext) Interface(key string, val interface{}) *LoggerContext {
	lc.zc = lc.zc.Interface(key, val)
	return lc
}

// Fields adds multiple fields from a map.
func (lc *LoggerContext) Fields(fields map[string]interface{}) *LoggerContext {
	lc.zc = lc.zc.Fields(fields)
	return lc
}

// TenantID adds a tenant_id field.
func (lc *LoggerContext) TenantID(tenantID string) *LoggerContext {
	lc.zc = lc.zc.Str("tenant_id", tenantID)
	return lc
}

// UserID adds a user_id field.
func (lc *LoggerContext) UserID(userID string) *LoggerContext {
	lc.zc = lc.zc.Str("user_id", userID)
	return lc
}

// RequestID adds a request_id field.
func (lc *LoggerContext) RequestID(requestID string) *LoggerContext {
	lc.zc = lc.zc.Str("request_id", requestID)
	return lc
}

// TraceID adds a trace_id field.
func (lc *LoggerContext) TraceID(traceID string) *LoggerContext {
	lc.zc = lc.zc.Str("trace_id", traceID)
	return lc
}

// SpanID adds a span_id field.
func (lc *LoggerContext) SpanID(spanID string) *LoggerContext {
	lc.zc = lc.zc.Str("span_id", spanID)
	return lc
}

// Service adds a service field.
func (lc *LoggerContext) Service(service string) *LoggerContext {
	lc.zc = lc.zc.Str("service", service)
	return lc
}

// Logger returns the configured logger.
func (lc *LoggerContext) Logger() *Logger {
	return &Logger{zl: lc.zc.Logger()}
}

// Log level methods

// Debug logs a debug message.
func (l *Logger) Debug() *Event {
	return &Event{ze: l.zl.Debug()}
}

// Info logs an info message.
func (l *Logger) Info() *Event {
	return &Event{ze: l.zl.Info()}
}

// Warn logs a warning message.
func (l *Logger) Warn() *Event {
	return &Event{ze: l.zl.Warn()}
}

// Error logs an error message.
func (l *Logger) Error() *Event {
	return &Event{ze: l.zl.Error()}
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal() *Event {
	return &Event{ze: l.zl.Fatal()}
}

// Panic logs a panic message and panics.
func (l *Logger) Panic() *Event {
	return &Event{ze: l.zl.Panic()}
}

// Event represents a log event.
type Event struct {
	ze *zerolog.Event
}

// Str adds a string field to the event.
func (e *Event) Str(key, val string) *Event {
	e.ze = e.ze.Str(key, val)
	return e
}

// Int adds an integer field to the event.
func (e *Event) Int(key string, val int) *Event {
	e.ze = e.ze.Int(key, val)
	return e
}

// Int64 adds an int64 field to the event.
func (e *Event) Int64(key string, val int64) *Event {
	e.ze = e.ze.Int64(key, val)
	return e
}

// Float64 adds a float64 field to the event.
func (e *Event) Float64(key string, val float64) *Event {
	e.ze = e.ze.Float64(key, val)
	return e
}

// Bool adds a boolean field to the event.
func (e *Event) Bool(key string, val bool) *Event {
	e.ze = e.ze.Bool(key, val)
	return e
}

// Time adds a time field to the event.
func (e *Event) Time(key string, val time.Time) *Event {
	e.ze = e.ze.Time(key, val)
	return e
}

// Dur adds a duration field to the event.
func (e *Event) Dur(key string, val time.Duration) *Event {
	e.ze = e.ze.Dur(key, val)
	return e
}

// Err adds an error field to the event.
func (e *Event) Err(err error) *Event {
	e.ze = e.ze.Err(err)
	return e
}

// Interface adds an interface field to the event.
func (e *Event) Interface(key string, val interface{}) *Event {
	e.ze = e.ze.Interface(key, val)
	return e
}

// Fields adds multiple fields from a map to the event.
func (e *Event) Fields(fields map[string]interface{}) *Event {
	e.ze = e.ze.Fields(fields)
	return e
}

// Stack adds a stack trace to the event.
func (e *Event) Stack() *Event {
	e.ze = e.ze.Stack()
	return e
}

// Msg sends the event with the given message.
func (e *Event) Msg(msg string) {
	e.ze.Msg(msg)
}

// Msgf sends the event with a formatted message.
func (e *Event) Msgf(format string, args ...interface{}) {
	e.ze.Msgf(format, args...)
}

// Send sends the event without a message.
func (e *Event) Send() {
	e.ze.Send()
}

// Convenience functions using the global logger

// Debug logs a debug message using the global logger.
func Debug() *Event {
	return globalLogger.Debug()
}

// Info logs an info message using the global logger.
func Info() *Event {
	return globalLogger.Info()
}

// Warn logs a warning message using the global logger.
func Warn() *Event {
	return globalLogger.Warn()
}

// Error logs an error message using the global logger.
func Error() *Event {
	return globalLogger.Error()
}

// Fatal logs a fatal message using the global logger and exits.
func Fatal() *Event {
	return globalLogger.Fatal()
}

// Panic logs a panic message using the global logger and panics.
func Panic() *Event {
	return globalLogger.Panic()
}

// With returns a new context builder using the global logger.
func With() *LoggerContext {
	return globalLogger.With()
}
