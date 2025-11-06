// File update_test.go exercises capture scheduling and helper sizing.
package ui

import (
	"fmt"
	"strings"
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

// TestPaneContentRespectsManualScroll keeps manual offsets when the user scrolls away from the bottom.
func TestPaneContentRespectsManualScroll(t *testing.T) {
	t.Parallel()

	lines := make([]string, 16)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%02d", i)
	}
	content := strings.Join(lines, "\n")
	vp := viewportFor(innerDimension{width: 80, height: 6})
	vp.SetContent(content)
	vp.GotoBottom()
	vp.ScrollUp(3)
	if vp.AtBottom() {
		t.Fatalf("expected viewport not at bottom after scrolling up")
	}
	preview := &sessionPreview{
		viewport:    &vp,
		paneID:      "%1",
		lastContent: content,
		lastChanged: time.Now(),
	}
	m := &Model{
		previews: map[string]*sessionPreview{
			"$1": preview,
		},
		stale: make(map[string]struct{}),
	}
	originalOffset := preview.viewport.YOffset()

	msg := paneContentMsg{
		sessionID: "$1",
		paneID:    "%1",
		text:      content + "\nextra-line",
	}
	m.Update(msg)

	if got := preview.viewport.YOffset(); got != originalOffset {
		t.Fatalf("unexpected YOffset, got %d want %d", got, originalOffset)
	}
	if preview.viewport.AtBottom() {
		t.Fatalf("viewport jumped to bottom after manual scroll")
	}
	wantContent := strings.TrimRight(msg.text, "\n")
	if got := preview.lastContent; got != wantContent {
		t.Fatalf("lastContent = %q, want %q", got, wantContent)
	}
}

// TestPaneContentFollowsBottom continues auto-follow when the viewport was already at the bottom.
func TestPaneContentFollowsBottom(t *testing.T) {
	t.Parallel()

	lines := make([]string, 12)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%02d", i)
	}
	content := strings.Join(lines, "\n")
	vp := viewportFor(innerDimension{width: 80, height: 6})
	vp.SetContent(content)
	vp.GotoBottom()
	if !vp.AtBottom() {
		t.Fatalf("expected viewport to be at bottom initially")
	}

	preview := &sessionPreview{
		viewport:    &vp,
		paneID:      "%1",
		lastContent: content,
		lastChanged: time.Now(),
	}
	m := &Model{
		previews: map[string]*sessionPreview{
			"$1": preview,
		},
		stale: make(map[string]struct{}),
	}

	msg := paneContentMsg{
		sessionID: "$1",
		paneID:    "%1",
		text:      content + "\nextra-line",
	}
	m.Update(msg)

	if !preview.viewport.AtBottom() {
		t.Fatalf("expected viewport to remain at bottom when already there")
	}
	wantContent := strings.TrimRight(msg.text, "\n")
	if got := preview.lastContent; got != wantContent {
		t.Fatalf("lastContent = %q, want %q", got, wantContent)
	}
}
