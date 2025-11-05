// File util.go collects small helpers that support layout and formatting.
package ui

import (
	"fmt"
	"time"
)

// coarseDuration rounds durations into coarse bins for display.
func coarseDuration(d time.Duration) string {
	switch {
	case d < 5*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()/5)*5)
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}

// max returns the largest of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
