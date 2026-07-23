package logging

import (
	"testing"
)

func TestLoggerCreation(t *testing.T) {
	logger, err := NewLogger("INFO", "console")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}
}

func TestComponentLogging(t *testing.T) {
	logger, _ := NewLogger("INFO", "console")
	configLogger := logger.With("config")
	if configLogger.component != "config" {
		t.Errorf("Expected component 'config', got: %s", configLogger.component)
	}
}
