// File sessions_test.go covers helpers that navigate tmux snapshot data.
package ui

import (
	"testing"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestSessionMatches ensures fuzzy queries detect sessions, windows, and panes.
func TestSessionMatches(t *testing.T) {
	t.Parallel()

	session := tmux.Session{
		Name: "Work",
		Windows: []tmux.Window{
			{
				Name: "build",
				Panes: []tmux.Pane{
					{Title: "logs"},
				},
			},
		},
	}

	tests := []struct {
		query string
		want  bool
	}{
		{query: "work", want: true},
		{query: "build", want: true},
		{query: "logs", want: true},
		{query: "missing", want: false},
	}

	for _, tt := range tests {
		if got := sessionMatches(session, tt.query); got != tt.want {
			t.Fatalf("sessionMatches(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

// TestActiveWindow confirms the active window is chosen correctly.
func TestActiveWindow(t *testing.T) {
	t.Parallel()

	windows := []tmux.Window{
		{Name: "first"},
		{Name: "second", Active: true},
	}

	session := tmux.Session{Windows: windows}
	got, ok := activeWindow(session)
	if !ok {
		t.Fatal("activeWindow returned false")
	}
	if got.Name != "second" {
		t.Fatalf("activeWindow = %s, want second", got.Name)
	}
}

// TestActivePane confirms pane selection prioritises the active flag.
func TestActivePane(t *testing.T) {
	t.Parallel()

	window := tmux.Window{
		Panes: []tmux.Pane{
			{ID: "%1"},
			{ID: "%2", Active: true},
		},
	}
	got, ok := activePane(window)
	if !ok {
		t.Fatal("activePane returned false")
	}
	if got.ID != "%2" {
		t.Fatalf("activePane = %s, want %%2", got.ID)
	}
}

// TestSessionAllPanesDead verifies detection of sessions with only dead panes.
func TestSessionAllPanesDead(t *testing.T) {
	t.Parallel()

	session := tmux.Session{
		Windows: []tmux.Window{
			{Panes: []tmux.Pane{{Dead: true}, {Dead: true}}},
		},
	}
	if !sessionAllPanesDead(session) {
		t.Fatal("expected session to be considered dead")
	}

	session.Windows[0].Panes[1].Dead = false
	if sessionAllPanesDead(session) {
		t.Fatal("expected session to be considered alive")
	}
}

// TestSessionLatestActivity returns the newest timestamp among pane activity.
func TestSessionLatestActivity(t *testing.T) {
	t.Parallel()

	now := time.Now()
	session := tmux.Session{
		CreatedAt: now.Add(-2 * time.Hour),
		Windows: []tmux.Window{
			{Panes: []tmux.Pane{{LastActivity: now.Add(-30 * time.Minute)}, {LastActivity: now.Add(-10 * time.Minute)}}},
		},
	}

	got := sessionLatestActivity(session)
	if !got.Equal(now.Add(-10 * time.Minute)) {
		t.Fatalf("sessionLatestActivity = %v, want %v", got, now.Add(-10*time.Minute))
	}
}

// TestSessionLatestActivityNoTimestamps ensures zero time when panes lack data.
func TestSessionLatestActivityNoTimestamps(t *testing.T) {
	t.Parallel()

	session := tmux.Session{
		Windows: []tmux.Window{{Panes: []tmux.Pane{{}, {}}}},
	}

	if got := sessionLatestActivity(session); !got.IsZero() {
		t.Fatalf("sessionLatestActivity should be zero when panes lack activity, got %v", got)
	}
}
