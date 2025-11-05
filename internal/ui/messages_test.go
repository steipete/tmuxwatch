// File messages_test.go exercises toast and formatting helpers.
package ui

import (
	"strings"
	"testing"
	"time"
)

// TestCenterText verifies helper output is centred within the requested width.
func TestCenterText(t *testing.T) {
	t.Parallel()

	got := centerText("hi", 10)
	if !strings.HasPrefix(got, strings.Repeat(" ", 4)) {
		t.Fatalf("centerText produced wrong left padding: %q", got)
	}
	if !strings.HasSuffix(got, "hi") {
		t.Fatalf("centerText should end with original text, got %q", got)
	}
}

// TestToastViewActive makes sure active toasts render and persist until expiry.
func TestToastViewActive(t *testing.T) {
	t.Parallel()

	var m Model
	m.showToast("session closed")
	out := m.toastView(20)
	if !strings.Contains(out, "session closed") {
		t.Fatalf("toastView should render toast text, got %q", out)
	}
	if m.toast == nil || m.toast.text == "" {
		t.Fatal("toast should remain active after rendering")
	}
}

// TestToastViewExpired confirms expired toasts disappear and reset state.
func TestToastViewExpired(t *testing.T) {
	t.Parallel()

	m := Model{toast: &toastState{text: "stale", exp: time.Now().Add(-time.Second)}}
	out := m.toastView(20)
	if out != "" {
		t.Fatalf("toastView should return empty after expiration, got %q", out)
	}
	if m.toast.text != "" {
		t.Fatal("toast text should be cleared once expired")
	}
}
