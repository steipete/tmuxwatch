// File sessions.go offers helpers for navigating through snapshot structures.
package ui

import (
	"strings"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// sessionMatches reports whether a session, its windows, or panes contain the
// provided query string.
func sessionMatches(session tmux.Session, query string) bool {
	if strings.Contains(strings.ToLower(session.Name), query) {
		return true
	}
	for _, window := range session.Windows {
		if strings.Contains(strings.ToLower(window.Name), query) {
			return true
		}
		for _, pane := range window.Panes {
			if strings.Contains(strings.ToLower(pane.TitleOrCmd()), query) {
				return true
			}
		}
	}
	return false
}

// activeWindow picks the active window for a session or falls back to the
// first window in the slice.
func activeWindow(session tmux.Session) (tmux.Window, bool) {
	for _, window := range session.Windows {
		if window.Active {
			return window, true
		}
	}
	if len(session.Windows) > 0 {
		return session.Windows[0], true
	}
	return tmux.Window{}, false
}

// activePane selects the active pane, defaulting to the first pane when none
// are marked active.
func activePane(window tmux.Window) (tmux.Pane, bool) {
	for _, pane := range window.Panes {
		if pane.Active {
			return pane, true
		}
	}
	if len(window.Panes) > 0 {
		return window.Panes[0], true
	}
	return tmux.Pane{}, false
}

// sessionAllPanesDead reports whether every pane in the session has exited.
func sessionAllPanesDead(session tmux.Session) bool {
	found := false
	for _, window := range session.Windows {
		for _, pane := range window.Panes {
			found = true
			if !pane.Dead {
				return false
			}
		}
	}
	return found
}

// sessionLatestActivity returns the most recent activity timestamp within a
// session.
func sessionLatestActivity(session tmux.Session) time.Time {
	var latest time.Time
	for _, window := range session.Windows {
		for _, pane := range window.Panes {
			ts := pane.LastActivity
			if ts.After(latest) {
				latest = ts
			}
		}
	}
	return latest
}
