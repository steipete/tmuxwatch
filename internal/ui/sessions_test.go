package ui

import (
	"testing"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

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
