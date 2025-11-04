package tmux

import (
	"testing"
	"time"
)

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

func TestParseUnixInvalid(t *testing.T) {
	t.Parallel()

	if _, err := parseUnix("abc"); err == nil {
		t.Fatal("parseUnix expected error for invalid input")
	}
}
