package discovery

import (
	"errors"
	"testing"
)

func TestWalkEvent_String(t *testing.T) {
	tests := []struct {
		event    WalkEvent
		expected string
	}{
		{EventEnterDir, "EnterDir"},
		{EventFoundRepo, "FoundRepo"},
		{EventError, "Error"},
		{EventSkipped, "Skipped"},
		{WalkEvent(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.event.String(); got != tt.expected {
				t.Errorf("WalkEvent.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiscoveryError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DiscoveryError
		contains string
	}{
		{
			name: "error with directory path",
			err: &DiscoveryError{
				RootPath: "/root",
				DirPath:  "/root/somedir",
				Err:      errors.New("test error"),
			},
			contains: "/root/somedir",
		},
		{
			name: "error without directory path",
			err: &DiscoveryError{
				RootPath: "/root",
				Err:      errors.New("test error"),
			},
			contains: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !contains(msg, tt.contains) {
				t.Errorf("Error() message = %v, should contain %v", msg, tt.contains)
			}
		})
	}
}

func TestDiscoveryError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	discErr := &DiscoveryError{
		RootPath: "/root",
		DirPath:  "/root/dir",
		Err:      originalErr,
	}

	unwrapped := discErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestConcurrentDiscoveryError_Error(t *testing.T) {
	err := &ConcurrentDiscoveryError{
		CurrentOperation: "DiscoverProjects",
	}

	msg := err.Error()
	if !contains(msg, "DiscoverProjects") {
		t.Errorf("Error() message = %v, should contain 'DiscoverProjects'", msg)
	}
	if !contains(msg, "concurrent") {
		t.Errorf("Error() message = %v, should contain 'concurrent'", msg)
	}
}

func TestConcurrentDiscoveryError_Unwrap(t *testing.T) {
	err := &ConcurrentDiscoveryError{
		CurrentOperation: "DiscoverProjects",
	}

	unwrapped := err.Unwrap()
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestDiscoveryResult(t *testing.T) {
	// Test creation and initialization
	result := &DiscoveryResult{
		Discovered: 5,
		Errors:     []DiscoveryError{},
		RootStats:  make(map[string]RootStat),
	}

	if result.Discovered != 5 {
		t.Errorf("expected Discovered = 5, got %d", result.Discovered)
	}

	if result.Errors == nil {
		t.Error("expected Errors to be initialized")
	}

	if result.RootStats == nil {
		t.Error("expected RootStats to be initialized")
	}

	// Test adding stats
	result.RootStats["/root1"] = RootStat{
		Discovered: 3,
		Errors:     1,
		DurationMs: 100,
	}

	if len(result.RootStats) != 1 {
		t.Errorf("expected 1 root stat, got %d", len(result.RootStats))
	}

	stat := result.RootStats["/root1"]
	if stat.Discovered != 3 {
		t.Errorf("expected Discovered = 3, got %d", stat.Discovered)
	}
}

func TestRootStat(t *testing.T) {
	stat := RootStat{
		Discovered: 10,
		Errors:     2,
		DurationMs: 500,
	}

	if stat.Discovered != 10 {
		t.Errorf("expected Discovered = 10, got %d", stat.Discovered)
	}
	if stat.Errors != 2 {
		t.Errorf("expected Errors = 2, got %d", stat.Errors)
	}
	if stat.DurationMs != 500 {
		t.Errorf("expected DurationMs = 500, got %d", stat.DurationMs)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
