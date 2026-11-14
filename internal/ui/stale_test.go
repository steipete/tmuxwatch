// File stale_test.go checks stale detection logic and activity tracking.
package ui

import (
	"testing"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestSessionActivityPrefersSessionTimestamp ensures session timestamps win unless newer preview data exists.
func TestSessionActivityPrefersSessionTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()
	session := tmux.Session{
		ID:           "s1",
		LastActivity: now.Add(-5 * time.Minute),
		Windows: []tmux.Window{
			{Panes: []tmux.Pane{{LastActivity: now.Add(-30 * time.Minute)}}},
		},
	}
	m := &Model{previews: make(map[string]*sessionPreview)}
	got := m.sessionActivity(session)
	if !got.Equal(session.LastActivity) {
		t.Fatalf("sessionActivity = %v, want %v", got, session.LastActivity)
	}

	m.previews["s1"] = &sessionPreview{lastChanged: now}
	got = m.sessionActivity(session)
	if got.Before(now) {
		t.Fatalf("sessionActivity should respect preview change, got %v", got)
	}
}

func TestFirstPreviewSampleDoesNotClearStale(t *testing.T) {
	t.Parallel()

	old := time.Now().Add(-2 * staleThreshold)
	session := tmux.Session{
		ID:           "s1",
		Name:         "idle",
		LastActivity: old,
		Windows: []tmux.Window{{
			Active: true,
			Panes:  []tmux.Pane{{ID: "%1", Active: true, LastActivity: old}},
		}},
	}
	m := &Model{
		sessions: []tmux.Session{session},
		previews: make(map[string]*sessionPreview),
		stale:    make(map[string]struct{}),
		width:    80,
		height:   24,
	}
	m.ensurePreviewsAndCapture()
	if !m.previews["s1"].lastChanged.IsZero() {
		t.Fatal("creating a preview restarted the stale clock")
	}
	_, _ = m.Update(paneContentMsg{sessionID: "s1", paneID: "%1", text: "prompt $\n"})
	m.updateStaleSessions()
	if !m.isStale("s1") {
		t.Fatal("first capture of existing text cleared stale")
	}
	_, _ = m.Update(paneContentMsg{sessionID: "s1", paneID: "%1", text: "prompt $\nnew line\n"})
	if m.isStale("s1") {
		t.Fatal("later output should clear stale")
	}
}

func TestEmptyPreviewBaselineTracksLaterOutput(t *testing.T) {
	t.Parallel()
	old := time.Now().Add(-2 * staleThreshold)
	m := &Model{
		sessions: []tmux.Session{{ID: "s1", LastActivity: old}},
		previews: make(map[string]*sessionPreview),
		stale:    make(map[string]struct{}),
	}
	vp := viewportFor(innerDimension{width: 40, height: 8})
	m.previews["s1"] = &sessionPreview{viewport: &vp, paneID: "%1"}
	_, _ = m.Update(paneContentMsg{sessionID: "s1", paneID: "%1", text: "\n"})
	m.updateStaleSessions()
	if !m.isStale("s1") || !m.previews["s1"].sampled {
		t.Fatal("empty baseline must stay stale and mark sampling complete")
	}
	_, _ = m.Update(paneContentMsg{sessionID: "s1", paneID: "%1", text: "new output\n"})
	if m.isStale("s1") {
		t.Fatal("output after an empty baseline must count as activity")
	}
}

// TestUpdateStaleSessionsClearsAfterActivity verifies stale sessions reactivate when new output arrives.
func TestUpdateStaleSessionsClearsAfterActivity(t *testing.T) {
	t.Parallel()

	old := time.Now().Add(-2 * staleThreshold)
	session := tmux.Session{
		ID:           "s1",
		Name:         "work",
		LastActivity: old,
		Windows: []tmux.Window{
			{Panes: []tmux.Pane{{LastActivity: old}}},
		},
	}
	m := &Model{
		sessions: []tmux.Session{session},
		previews: map[string]*sessionPreview{
			"s1": {lastChanged: old},
		},
		stale: make(map[string]struct{}),
	}

	m.updateStaleSessions()
	if !m.isStale("s1") {
		t.Fatal("expected session to be marked stale")
	}

	m.previews["s1"].lastChanged = time.Now()
	m.updateStaleSessions()
	if m.isStale("s1") {
		t.Fatal("expected session to become active after new content")
	}
}
