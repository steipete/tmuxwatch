package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestVersionFlag verifies --version outputs the version string and exits cleanly
func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Args = append(cmd.Args, "--version")
	cmd.Env = os.Environ()

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/tmuxwatch-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer os.Remove("/tmp/tmuxwatch-test")

	cmd = exec.Command("/tmp/tmuxwatch-test", "--version")
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

	// Build the binary once
	buildCmd := exec.Command("go", "build", "-o", "/tmp/tmuxwatch-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer os.Remove("/tmp/tmuxwatch-test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("/tmp/tmuxwatch-test", "--debug-click", tt.value)
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
	buildCmd := exec.Command("go", "build", "-o", "/tmp/tmuxwatch-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer os.Remove("/tmp/tmuxwatch-test")

	cmd := exec.Command("/tmp/tmuxwatch-test", "--interval", "invalid")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("expected error for invalid --interval value")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "invalid") {
		t.Errorf("expected error message about invalid interval, got: %s", outputStr)
	}
}

// TestHelpFlag verifies -h/--help flags work
func TestHelpFlag(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "/tmp/tmuxwatch-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer os.Remove("/tmp/tmuxwatch-test")

	cmd := exec.Command("/tmp/tmuxwatch-test", "-h")
	output, err := cmd.CombinedOutput()

	// -h exits with status 0 in flag package
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage of") && !strings.Contains(outputStr, "interval") {
		t.Errorf("expected help output to contain usage information, got: %s", outputStr)
	}
	_ = err // May or may not error depending on flag package behavior
}
