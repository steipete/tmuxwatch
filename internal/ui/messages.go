// File messages.go owns transient status/toast helpers that decorate the
// footer.
package ui

import (
	"strings"
	"time"
)

// toastState tracks the current toast message and its expiration time.
type toastState struct {
	text string
	exp  time.Time
}

// showToast updates the transient toast message and schedules its expiration.
func (m *Model) showToast(msg string) {
	if m.toast == nil {
		m.toast = &toastState{}
	}
	m.toast.text = msg
	m.toast.exp = time.Now().Add(3 * time.Second)
}

// toastView renders the toast centred on the footer or returns an empty string
// when the toast is inactive or expired.
func (m *Model) toastView(width int) string {
	if m.toast == nil || m.toast.text == "" {
		return ""
	}
	if time.Now().After(m.toast.exp) {
		m.toast.text = ""
		return ""
	}
	return centerText(m.toast.text, width)
}

// centerText pads text with spaces so it appears centred for the requested
// width.
func centerText(text string, width int) string {
	if width <= 0 {
		width = len(text)
	}
	if len(text) >= width {
		return text
	}
	pad := (width - len(text)) / 2
	return strings.Repeat(" ", pad) + text
}
