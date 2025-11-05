package ui

import (
	"testing"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

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
