package logging

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"project-dash/pkg/models"
)

func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		wantErr bool
	}{
		{"info console", "INFO", "console", false},
		{"debug json", "DEBUG", "json", false},
		{"invalid level", "INVALID", "console", true},
		{"invalid format", "INFO", "invalid", false}, // defaults to console
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := NewLogger("DEBUG", "console")

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify all messages were logged
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not found in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not found in output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := NewLogger("INFO", "console")

	logger.Debug("debug message")
	logger.Info("info message")

	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Debug should not appear, info should
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear at INFO level")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found in output")
	}
}

func TestComponentLogging(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := NewLogger("DEBUG", "json")
	configLogger := logger.With("config")

	configLogger.Info("config loaded",
		models.Field{Key: "path", Value: "/test/config.toml"},
	)

	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify component field exists
	if !strings.Contains(output, "\"component\":\"config\"") {
		t.Error("Component field not found in JSON output")
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("PORTFOLIO_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("PORTFOLIO_LOG_LEVEL")

	level := GetLogLevelFromEnv()
	if level != "DEBUG" {
		t.Errorf("Expected DEBUG from env, got: %s", level)
	}
}

func TestLoggerSync(t *testing.T) {
	// Capture stdout to avoid interfering with other tests
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := NewLogger("INFO", "console")

	logger.Info("test message")

	// Close the writer first to flush all content
	w.Close()
	os.Stdout = old

	// Read any remaining output to avoid blocking
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Sync should not error even after closing stdout
	if err := logger.Sync(); err != nil {
		// Zap returns a specific error when syncing to closed writer, which is expected
		if err.Error() != "sync: write on closed pipe" && !errors.Is(err, os.ErrClosed) {
			t.Errorf("Sync() failed unexpectedly: %v", err)
		}
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test global logger initialization
	err := InitializeGlobalLogger("INFO", "console")
	if err != nil {
		t.Errorf("Failed to initialize global logger: %v", err)
	}

	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("Global logger should not be nil")
	}

	// Test default initialization
	globalLogger = nil // Reset
	logger2 := GetGlobalLogger()
	if logger2 == nil {
		t.Error("Default global logger should not be nil")
	}
}

func TestLoggerWithComponent(t *testing.T) {
	logger, _ := NewLogger("INFO", "console")

	// Test With method
	configLogger := logger.With("config")
	if configLogger.component != "config" {
		t.Errorf("Expected component 'config', got: %s", configLogger.component)
	}

	// Test that original logger is unchanged
	if logger.component != "" {
		t.Error("Original logger should not have component set")
	}
}

func TestLogFormatValidation(t *testing.T) {
	config := &Config{
		Level:  "INFO",
		Format: "json",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test invalid format
	config2 := &Config{
		Level:  "INFO",
		Format: "invalid",
	}

	err2 := config2.Validate()
	if err2 == nil {
		t.Error("Invalid format should fail validation")
	}
}

func TestConfigLoading(t *testing.T) {
	// Test loading config from environment
	os.Setenv("PORTFOLIO_LOG_LEVEL", "DEBUG")
	os.Setenv("PORTFOLIO_LOG_FORMAT", "json")
	defer os.Unsetenv("PORTFOLIO_LOG_LEVEL")
	defer os.Unsetenv("PORTFOLIO_LOG_FORMAT")

	config := LoadConfigFromEnv()
	if config.Level != "DEBUG" {
		t.Errorf("Expected DEBUG level, got: %s", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected json format, got: %s", config.Format)
	}
}

func TestGetEffectiveLevel(t *testing.T) {
	config := &Config{
		Level:  "INFO",
		Format: "console",
		Components: map[string]string{
			"database": "DEBUG",
		},
	}

	// Test component-specific level
	dbLevel := config.GetEffectiveLevel("database")
	if dbLevel != "DEBUG" {
		t.Errorf("Expected DEBUG for database component, got: %s", dbLevel)
	}

	// Test default level
	defaultLevel := config.GetEffectiveLevel("config")
	if defaultLevel != "INFO" {
		t.Errorf("Expected INFO for non-specified component, got: %s", defaultLevel)
	}
}

func TestFieldLogging(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := NewLogger("DEBUG", "json")

	logger.Info("test with fields",
		models.Field{Key: "key1", Value: "value1"},
		models.Field{Key: "key2", Value: 42},
		models.Field{Key: "key3", Value: true},
	)

	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify field values exist
	if !strings.Contains(output, "\"key1\":\"value1\"") {
		t.Error("Field key1 not found in output")
	}
	if !strings.Contains(output, "\"key2\":42") {
		t.Error("Field key2 not found in output")
	}
	if !strings.Contains(output, "\"key3\":true") {
		t.Error("Field key3 not found in output")
	}
}

// TestStory13AcceptanceCriteria tests acceptance criteria for Story 1.3
func TestStory13AcceptanceCriteria(t *testing.T) {
	t.Run("AC10_StructuredLogging", func(t *testing.T) {
		// Test structured logging with JSON format
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		logger, err := NewLogger("INFO", "json")
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Info("structured message",
			models.Field{Key: "component", Value: "test"},
			models.Field{Key: "action", Value: "test_action"},
		)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify JSON structure
		if !strings.Contains(output, "\"message\":\"structured message\"") {
			t.Error("Message field not found in JSON output")
		}
		if !strings.Contains(output, "\"component\":\"test\"") {
			t.Error("Component field not found in JSON output")
		}
	})

	t.Run("AC11_LogLevels", func(t *testing.T) {
		// Test all log levels work
		levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
		for _, level := range levels {
			logger, err := NewLogger(level, "console")
			if err != nil {
				t.Errorf("Failed to create logger with level %s: %v", level, err)
			}
			if logger == nil {
				t.Errorf("Logger is nil for level %s", level)
			}
		}
	})

	t.Run("AC12_StandardOutput", func(t *testing.T) {
		// Test logging to standard output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		logger, _ := NewLogger("INFO", "console")
		logger.Info("stdout test")

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "stdout test") {
			t.Error("Log output not found in stdout")
		}
	})

	t.Run("AC13_EnvironmentConfiguration", func(t *testing.T) {
		// Test environment variable configuration
		os.Setenv("PORTFOLIO_LOG_LEVEL", "DEBUG")
		os.Setenv("PORTFOLIO_LOG_FORMAT", "json")
		defer func() {
			os.Unsetenv("PORTFOLIO_LOG_LEVEL")
			os.Unsetenv("PORTFOLIO_LOG_FORMAT")
		}()

		config := LoadConfigFromEnv()

		if config.Level != "DEBUG" {
			t.Errorf("Expected DEBUG from env, got: %s", config.Level)
		}
		if config.Format != "json" {
			t.Errorf("Expected json from env, got: %s", config.Format)
		}
	})
}
