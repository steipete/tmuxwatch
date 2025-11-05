// File layout_test.go verifies cursor management within the grid layout.
package ui

import (
	"testing"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestEnsureCursor ensures the cursor tracks visible sessions correctly.
func TestEnsureCursor(t *testing.T) {
	t.Parallel()

	sessions := []tmux.Session{{ID: "$1"}, {ID: "$2"}}
	m := &Model{hidden: make(map[string]struct{}), cursorSession: "$missing"}
	m.ensureCursor(sessions)
	if m.cursorSession != "$1" {
		t.Fatalf("cursor fallback = %q, want $1", m.cursorSession)
	}

	m.cursorSession = "$2"
	m.ensureCursor(sessions)
	if m.cursorSession != "$2" {
		t.Fatalf("cursor should stay when valid, got %q", m.cursorSession)
	}

	m.ensureCursor(nil)
	if m.cursorSession != "" {
		t.Fatalf("cursor should clear when no sessions, got %q", m.cursorSession)
	}
}

// TestMoveCursorGrid exercises multi-column cursor navigation helpers.
func TestMoveCursorGrid(t *testing.T) {
	t.Parallel()

	sessions := []tmux.Session{{ID: "$1"}, {ID: "$2"}, {ID: "$3"}, {ID: "$4"}}
	m := &Model{
		sessions:      sessions,
		hidden:        make(map[string]struct{}),
		cursorSession: "$1",
		cardCols:      2,
	}

	if !m.moveCursorRight() {
		t.Fatal("expected move right from first column")
	}
	if m.cursorSession != "$2" {
		t.Fatalf("cursor = %q, want $2", m.cursorSession)
	}

	if m.moveCursorRight() {
		t.Fatal("should not wrap rows when moving right")
	}
	if m.cursorSession != "$2" {
		t.Fatalf("cursor should remain $2, got %q", m.cursorSession)
	}

	if !m.moveCursorDown() {
		t.Fatal("expected move down into next row")
	}
	if m.cursorSession != "$4" {
		t.Fatalf("cursor = %q, want $4", m.cursorSession)
	}

	if !m.moveCursorLeft() {
		t.Fatal("left should move within the same row")
	}
	if m.cursorSession != "$3" {
		t.Fatalf("cursor = %q, want $3", m.cursorSession)
	}

	if !m.moveCursorUp() {
		t.Fatal("expected move up to previous row")
	}
	if m.cursorSession != "$1" {
		t.Fatalf("cursor = %q, want $1", m.cursorSession)
	}
}
