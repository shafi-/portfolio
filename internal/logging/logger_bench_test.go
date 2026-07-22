package logging

import (
	"testing"

	"github.com/nerddevsltd/portfolio/pkg/models"
)

func BenchmarkLogger(b *testing.B) {
	logger, _ := NewLogger("INFO", "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			models.Field{Key: "iteration", Value: i},
		)
	}
	logger.Sync()
}

func BenchmarkLoggerWithField(b *testing.B) {
	logger, _ := NewLogger("INFO", "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			models.Field{Key: "iteration", Value: i},
			models.Field{Key: "component", Value: "test"},
			models.Field{Key: "timestamp", Value: "2024-01-01T00:00:00Z"},
		)
	}
	logger.Sync()
}

func BenchmarkLoggerConsole(b *testing.B) {
	logger, _ := NewLogger("INFO", "console")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			models.Field{Key: "iteration", Value: i},
		)
	}
	logger.Sync()
}

func BenchmarkLoggerWithComponent(b *testing.B) {
	logger, _ := NewLogger("INFO", "json")
	configLogger := logger.With("config")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configLogger.Info("benchmark message",
			models.Field{Key: "iteration", Value: i},
		)
	}
	logger.Sync()
}

func BenchmarkLogLevelDebug(b *testing.B) {
	logger, _ := NewLogger("DEBUG", "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message",
			models.Field{Key: "value", Value: i},
		)
	}
	logger.Sync()
}

func BenchmarkLoggerAllLevels(b *testing.B) {
	logger, _ := NewLogger("DEBUG", "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")
	}
	logger.Sync()
}

func BenchmarkConfigValidation(b *testing.B) {
	config := &Config{
		Level:  "INFO",
		Format: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

func BenchmarkNewLogger(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger, _ := NewLogger("INFO", "console")
		logger.Sync()
	}
}