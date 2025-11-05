// File util_test.go validates miscellaneous utility helpers.
package ui

import (
	"testing"
	"time"
)

// TestCoarseDuration checks time values render with coarse human strings.
func TestCoarseDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{name: "just now", in: 2 * time.Second, want: "just now"},
		{name: "seconds rounded", in: 17 * time.Second, want: "15s"},
		{name: "minutes", in: 3 * time.Minute, want: "3m"},
		{name: "hours", in: 4 * time.Hour, want: "4h"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := coarseDuration(tt.in); got != tt.want {
				t.Fatalf("coarseDuration(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestMax ensures the helper returns the greater of two integers.
func TestMax(t *testing.T) {
	t.Parallel()

	if got := max(2, 10); got != 10 {
		t.Fatalf("max(2, 10) = %d, want 10", got)
	}
	if got := max(7, -1); got != 7 {
		t.Fatalf("max(7, -1) = %d, want 7", got)
	}
}

// TestMin ensures the helper returns the lesser of two integers.
func TestMin(t *testing.T) {
	t.Parallel()

	if got := min(2, 10); got != 2 {
		t.Fatalf("min(2, 10) = %d, want 2", got)
	}
	if got := min(7, -1); got != -1 {
		t.Fatalf("min(7, -1) = %d, want -1", got)
	}
}
