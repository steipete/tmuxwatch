// Package tmux provides a thin wrapper around the tmux binary so the UI can
// reason about sessions, windows, and panes without parsing shell output.
package tmux

import (
	"fmt"
	"strings"
	"time"
)

// Session represents a tmux session with its windows populated.
type Session struct {
	ID        string
	Name      string
	Attached  bool
	CreatedAt time.Time
	// LastActivity records the most recent activity timestamp reported by tmux.
	LastActivity time.Time
	Windows      []Window
}

// Window represents a tmux window and its panes.
type Window struct {
	ID       string
	Name     string
	Active   bool
	Session  string
	Index    int
	LastPane time.Time
	Panes    []Pane
}

// Pane represents a tmux pane.
type Pane struct {
	ID            string
	Title         string
	Active        bool
	Window        string
	Session       string
	CurrentCmd    string
	TTY           string
	LastActivity  time.Time
	CreatedAt     time.Time
	Width, Height int
	Dead          bool
	DeadStatus    int
}

// Snapshot contains the state of the tmux server.
type Snapshot struct {
	Sessions  []Session
	Timestamp time.Time
}

// TitleOrCmd returns the most descriptive label for a pane for display
// purposes, preferring the title and falling back to the running command.
func (p Pane) TitleOrCmd() string {
	title := strings.TrimSpace(p.Title)
	if title != "" {
		return title
	}
	cmd := strings.TrimSpace(p.CurrentCmd)
	if cmd != "" {
		return cmd
	}
	return "pane"
}

// StatusString reports whether the pane is still running or the exit code if
// it has terminated.
func (p Pane) StatusString() string {
	if !p.Dead {
		return "running"
	}
	if p.DeadStatus == 0 {
		return "exit 0"
	}
	return fmt.Sprintf("exit %d", p.DeadStatus)
}
