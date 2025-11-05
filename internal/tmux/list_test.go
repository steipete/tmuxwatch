// File list_test.go validates helper parsing in the tmux list routines.
package tmux

import (
	"testing"
	"time"
)

// TestParseUnix ensures numeric strings convert to time correctly.
func TestParseUnix(t *testing.T) {
	t.Parallel()

	ts := time.Unix(100, 0)
	got, err := parseUnix("100")
	if err != nil {
		t.Fatalf("parseUnix returned unexpected error: %v", err)
	}
	if !got.Equal(ts) {
		t.Fatalf("parseUnix() = %v, want %v", got, ts)
	}
}

// TestParseUnixEmpty confirms blank values return zero time.
func TestParseUnixEmpty(t *testing.T) {
	t.Parallel()

	got, err := parseUnix(" ")
	if err != nil {
		t.Fatalf("parseUnix returned unexpected error: %v", err)
	}
	if !got.IsZero() {
		t.Fatalf("parseUnix() = %v, want zero time", got)
	}
}

// TestParseUnixInvalid verifies invalid input errors out.
func TestParseUnixInvalid(t *testing.T) {
	t.Parallel()

	if _, err := parseUnix("abc"); err == nil {
		t.Fatal("parseUnix expected error for invalid input")
	}
}
