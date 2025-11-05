// File update_test.go exercises capture scheduling and helper sizing.
package ui

import (
	"testing"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestCaptureLinesFor clamps capture sizes to configured bounds.
func TestCaptureLinesFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		height int
		want   int
	}{
		{name: "zero height uses min", height: 0, want: minCaptureLines},
		{name: "small height adds slack", height: 20, want: minCaptureLines},
		{name: "large height caps", height: 1000, want: maxCaptureLines},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := captureLinesFor(tc.height); got != tc.want {
				t.Fatalf("captureLinesFor(%d) = %d, want %d", tc.height, got, tc.want)
			}
		})
	}
}

// TestEnsurePreviewsSkipsCollapsed avoids captures when cards are collapsed and unfocused.
func TestEnsurePreviewsSkipsCollapsed(t *testing.T) {
	t.Parallel()

	m := &Model{
		previews:  make(map[string]*sessionPreview),
		hidden:    make(map[string]struct{}),
		stale:     make(map[string]struct{}),
		collapsed: map[string]struct{}{"$1": {}},
		width:     120,
		height:    60,
	}
	m.sessions = []tmux.Session{{
		ID: "$1",
		Windows: []tmux.Window{{
			Active: true,
			Panes:  []tmux.Pane{{ID: "%1", Active: true, LastActivity: time.Now()}},
		}},
	}}

	if cmd := m.ensurePreviewsAndCapture(); cmd != nil {
		t.Fatalf("expected no capture command when session is collapsed and unfocused")
	}
}

// TestEnsurePreviewsCapturesDetailSessions keeps detail view content up to date even if collapsed.
func TestEnsurePreviewsCapturesDetailSessions(t *testing.T) {
	t.Parallel()

	m := &Model{
		previews:      make(map[string]*sessionPreview),
		hidden:        make(map[string]struct{}),
		stale:         make(map[string]struct{}),
		collapsed:     map[string]struct{}{"$1": {}},
		width:         120,
		height:        60,
		viewMode:      viewModeDetail,
		detailSession: "$1",
	}
	m.sessions = []tmux.Session{{
		ID: "$1",
		Windows: []tmux.Window{{
			Active: true,
			Panes:  []tmux.Pane{{ID: "%1", Active: true, LastActivity: time.Now()}},
		}},
	}}

	if cmd := m.ensurePreviewsAndCapture(); cmd == nil {
		t.Fatalf("expected capture command for detail session")
	}
}
