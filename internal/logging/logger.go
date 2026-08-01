package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"project-dash/internal/version"
	"project-dash/pkg/models"
)

// Logger provides structured logging capabilities
type Logger struct {
	zapLogger   *zap.Logger
	errorLogger *zap.Logger
	component   string
}

// global logger instance
var globalLogger *Logger

// NewLogger creates a new structured logger
func NewLogger(level string, format string) (*Logger, error) {
	return NewLoggerWithFile(level, format, "", os.Stdout)
}

// NewLoggerWithFile creates a logger that writes to both file and console.
// The file always captures INFO+ level (full trail for debugging), while
// console output respects the provided level (typically ERROR by default,
// INFO with --verbose). If filePath is empty, file logging is disabled.
func NewLoggerWithFile(level string, format string, filePath string, consoleWriter io.Writer) (*Logger, error) {
	return NewLoggerWithFiles(level, format, filePath, "", consoleWriter)
}

// NewLoggerWithFiles creates a logger that writes to console, regular log file,
// and optionally an error log file. The error log captures ERROR+ level with
// full stack traces for debugging.
func NewLoggerWithFiles(level string, format string, filePath string, errorFilePath string, consoleWriter io.Writer) (*Logger, error) {
	// Parse the console log level (user-facing, may be ERROR or INFO)
	consoleLevel, err := parseLogLevel(level)
	if err != nil {
		return nil, err
	}

	// Configure encoder for console
	var consoleEncoderConfig zapcore.EncoderConfig
	if format == "json" {
		consoleEncoderConfig = zapcore.EncoderConfig{
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
		}
	} else {
		// Clean, readable console format
		consoleEncoderConfig = zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"), // Short HH:MM:SS
			EncodeDuration: zapcore.StringDurationEncoder,
		}
	}

	// Configure encoder for file (plain text, no color)
	fileEncoderConfig := zapcore.EncoderConfig{
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
	}

	// Create encoders
	var consoleEncoder, fileEncoder zapcore.Encoder
	if format == "json" {
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
	}

	// Create console core
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(consoleWriter),
		consoleLevel,
	)

	// If no file path, return console-only logger
	if filePath == "" {
		zapLogger := zap.New(consoleCore,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
			zap.Fields(zap.String("version", version.Version())),
		)
		return &Logger{zapLogger: zapLogger}, nil
	}

	// Open log file with rotation (30-day retention, 100MB max size, compress old logs)
	logFile := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100,  // megabytes
		MaxBackups: 10,   // keep last 10 rotated files
		MaxAge:     30,   // days
		Compress:   true, // gzip old logs
	}

	// Create file core (always INFO level for debugging)
	fileCore := zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(logFile),
		zapcore.InfoLevel,
	)

	// Combine cores with tee: write to both console and file
	core := zapcore.NewTee(consoleCore, fileCore)

	// Create logger with version as a global field
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("version", version.Version())),
	)

	// Create error-specific logger if errorFilePath is provided
	var errorLogger *zap.Logger
	if errorFilePath != "" {
		errorLogFile := &lumberjack.Logger{
			Filename:   errorFilePath,
			MaxSize:    50, // megabytes
			MaxBackups: 5,  // keep last 5 rotated files
			MaxAge:     90, // days - keep error logs longer
			Compress:   true,
		}

		// Error log always captures ERROR+ with full stack traces
		errorEncoderConfig := zapcore.EncoderConfig{
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
		}

		var errorEncoder zapcore.Encoder
		if format == "json" {
			errorEncoder = zapcore.NewJSONEncoder(errorEncoderConfig)
		} else {
			errorEncoder = zapcore.NewConsoleEncoder(errorEncoderConfig)
		}

		errorFileCore := zapcore.NewCore(
			errorEncoder,
			zapcore.AddSync(errorLogFile),
			zapcore.ErrorLevel,
		)

		errorLogger = zap.New(errorFileCore,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
			zap.Fields(zap.String("version", version.Version())),
		)
	}

	return &Logger{zapLogger: zapLogger, errorLogger: errorLogger}, nil
}

// NewStderrLogger creates a logger that writes to stderr
// Used for MCP server to keep stdout clean for JSON-RPC messages
func NewStderrLogger(level string, format string) (*Logger, error) {
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
		}
		// Omit CallerKey from config to skip caller in output
	} else {
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
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"), // Short HH:MM:SS
			EncodeDuration: zapcore.StringDurationEncoder,
		}
		// Omit CallerKey from config to skip caller in output
	}

	// Create encoder
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core with stderr output
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stderr),
		zapLevel,
	)

	// Create logger with version as a global field
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("version", version.Version())),
	)

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

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// Initialize with defaults if not already initialized
		logger, _ := NewLogger("INFO", "console")
		globalLogger = logger
	}

	return globalLogger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Zap returns the underlying zap.Logger for use by components that require it
func (l *Logger) Zap() *zap.Logger {
	return l.zapLogger
}

// With creates a new logger with a component tag
func (l *Logger) With(component string) *Logger {
	return &Logger{
		zapLogger: l.zapLogger,
		component: component,
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

// LogErrorToFile logs an error with full details to the error log file
// This includes stack traces and internal details that should not be shown to users
func (l *Logger) LogErrorToFile(msg string, err error, fields ...models.Field) {
	if l.errorLogger == nil {
		return
	}

	// Add error field
	allFields := make([]models.Field, 0, len(fields)+1)
	allFields = append(allFields, models.Field{Key: "error", Value: err.Error()})
	allFields = append(allFields, fields...)

	// Convert to zap fields
	zapFields := make([]zap.Field, 0, len(allFields))
	for _, field := range allFields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}

	// Log to error file with full details
	l.errorLogger.Error(msg, zapFields...)
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
