package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"project-dash/pkg/models"
)

// Logger provides structured logging capabilities
type Logger struct {
	zapLogger *zap.Logger
	component string
	once      sync.Once
}

// global logger instance
var globalLogger *Logger
var globalMutex sync.Mutex

// NewLogger creates a new structured logger
func NewLogger(level string, format string) (*Logger, error) {
	// Parse log level
	zapLevel, err := parseLogLevel(level)
	if err != nil {
		return nil, err
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if format == "json" {
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	} else {
		// Human-readable format
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	// Create encoder
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{zapLogger: zapLogger}, nil
}

// NewLoggerWithComponent creates a logger with a component tag
func NewLoggerWithComponent(level string, format string, component string) (*Logger, error) {
	logger, err := NewLogger(level, format)
	if err != nil {
		return nil, err
	}
	logger.component = component
	return logger, nil
}

// parseLogLevel converts string level to zap level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zapcore.DebugLevel, nil
	case "INFO":
		return zapcore.InfoLevel, nil
	case "WARN", "WARNING":
		return zapcore.WarnLevel, nil
	case "ERROR":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// InitializeGlobalLogger creates the global logger instance
func InitializeGlobalLogger(level string, format string) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	logger, err := NewLogger(level, format)
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if globalLogger == nil {
		// Initialize with defaults if not already initialized
		logger, _ := NewLogger("INFO", "console")
		globalLogger = logger
	}

	return globalLogger
}

// With creates a new logger with a component tag
func (l *Logger) With(component string) *Logger {
	return &Logger{
		zapLogger: l.zapLogger,
		component: component,
		once:      sync.Once{}, // New instance, not copied
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...models.Field) {
	l.log(models.DEBUG, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...models.Field) {
	l.log(models.INFO, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...models.Field) {
	l.log(models.WARN, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...models.Field) {
	l.log(models.ERROR, msg, fields...)
}

// log performs the actual logging
func (l *Logger) log(level models.LogLevel, msg string, fields ...models.Field) {
	// Add component field if present
	zapFields := make([]zap.Field, 0, len(fields)+1)

	if l.component != "" {
		zapFields = append(zapFields, zap.String("component", l.component))
	}

	// Add custom fields
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}

	// Log at appropriate level
	switch level {
	case models.DEBUG:
		l.zapLogger.Debug(msg, zapFields...)
	case models.INFO:
		l.zapLogger.Info(msg, zapFields...)
	case models.WARN:
		l.zapLogger.Warn(msg, zapFields...)
	case models.ERROR:
		l.zapLogger.Error(msg, zapFields...)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// GetLogLevelFromEnv returns log level from environment variable or default
func GetLogLevelFromEnv() string {
	if level := os.Getenv("PORTFOLIO_LOG_LEVEL"); level != "" {
		return level
	}
	return "INFO"
}

// GetLogFormatFromEnv returns log format from environment variable or default
func GetLogFormatFromEnv() string {
	if format := os.Getenv("PORTFOLIO_LOG_FORMAT"); format != "" {
		return format
	}
	return "console"
}
