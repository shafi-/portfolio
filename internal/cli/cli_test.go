package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	rootCmd := GetRootCommand()

	// Test basic command properties
	if rootCmd.Use != "portfolio" {
		t.Errorf("Expected use 'portfolio', got: %s", rootCmd.Use)
	}

	if !strings.Contains(rootCmd.Short, "Portfolio Engine") {
		t.Error("Short description missing Portfolio Engine")
	}
}

func TestHelpText(t *testing.T) {
	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := GetRootCommand()
	rootCmd.SetArgs([]string{"--help"})
	rootCmd.Execute()

	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify help content
	if !strings.Contains(output, "init") {
		t.Error("Help text missing 'init' command")
	}

	if !strings.Contains(output, "status") {
		t.Error("Help text missing 'status' command")
	}

	if !strings.Contains(output, "doctor") {
		t.Error("Help text missing 'doctor' command")
	}
}

func TestStatusCommand(t *testing.T) {
	// Test status command exists
	rootCmd := GetRootCommand()
	statusCmd, _, err := rootCmd.Find([]string{"status"})

	if err != nil {
		t.Errorf("Status command not found: %v", err)
	}

	if statusCmd == nil {
		t.Error("Status command is nil")
	}
}

func TestDoctorCommand(t *testing.T) {
	// Test doctor command exists
	rootCmd := GetRootCommand()
	doctorCmd, _, err := rootCmd.Find([]string{"doctor"})

	if err != nil {
		t.Errorf("Doctor command not found: %v", err)
	}

	if doctorCmd == nil {
		t.Error("Doctor command is nil")
	}
}

func TestInitCommand(t *testing.T) {
	// Test init command exists
	rootCmd := GetRootCommand()
	initCmd, _, err := rootCmd.Find([]string{"init"})

	if err != nil {
		t.Errorf("Init command not found: %v", err)
	}

	if initCmd == nil {
		t.Error("Init command is nil")
	}
}