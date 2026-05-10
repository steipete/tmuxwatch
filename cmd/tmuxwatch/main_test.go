package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testBinPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "tmuxwatch-test-")
	if err != nil {
		log.Fatalf("failed to create temp dir for test binary: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testBinPath = filepath.Join(tmpDir, "tmuxwatch-test")

	buildCmd := exec.Command("go", "build", "-o", testBinPath, ".")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		log.Fatalf("failed to build binary: %v\nOutput: %s", err, string(output))
	}

	os.Exit(m.Run())
}

// TestVersionFlag verifies --version outputs the version string and exits cleanly
func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(testBinPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version flag failed: %v, output: %s", err, output)
	}

	outputStr := strings.TrimSpace(string(output))
	expected := "tmuxwatch " + version
	if outputStr != expected {
		t.Errorf("expected version output %q, got %q", expected, outputStr)
	}
}

// TestInvalidDebugClick verifies invalid --debug-click values are rejected
func TestInvalidDebugClick(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError string
	}{
		{
			name:        "single value",
			value:       "10",
			expectError: "invalid --debug-click value",
		},
		{
			name:        "three values",
			value:       "10,20,30",
			expectError: "invalid --debug-click value",
		},
		{
			name:        "invalid x coordinate",
			value:       "abc,20",
			expectError: "invalid debug-click x coordinate",
		},
		{
			name:        "invalid y coordinate",
			value:       "10,xyz",
			expectError: "invalid debug-click y coordinate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(testBinPath, "--debug-click", tt.value)
			output, err := cmd.CombinedOutput()

			if err == nil {
				t.Errorf("expected error for --debug-click=%q, got none", tt.value)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expectError) {
				t.Errorf("expected error containing %q, got: %s", tt.expectError, outputStr)
			}
		})
	}
}

// TestInvalidIntervalFlag verifies invalid --interval values are rejected
func TestInvalidIntervalFlag(t *testing.T) {
	cmd := exec.Command(testBinPath, "--interval", "invalid")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error for invalid --interval value")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "invalid value") {
		t.Errorf("expected error message about invalid value, got: %s", outputStr)
	}
}

// TestHelpFlag verifies -h/--help flags work
func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(testBinPath, "-h")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("-h flag failed: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage of") && !strings.Contains(outputStr, "interval") {
		t.Errorf("expected help output to contain usage information, got: %s", outputStr)
	}
}
