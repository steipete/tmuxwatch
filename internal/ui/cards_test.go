// File cards_test.go covers card header formatting logic.
package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestFormatHeaderOmitsHost ensures pane titles that match the host are hidden.
func TestFormatHeaderOmitsHost(t *testing.T) {
	t.Parallel()

	session := tmux.Session{Name: "s", Windows: []tmux.Window{{Name: "win"}}}
	window := session.Windows[0]
	pane := tmux.Pane{
		Title:        "dev-host",
		LastActivity: time.Now().Add(-time.Minute),
	}

	got := formatHeader(80, session, window, pane, false, false, false, false, "[x]", "dev-host")
	if strings.Contains(got, "dev-host") {
		t.Fatalf("formatHeader should omit host when title matches, got %q", got)
	}
}

// TestFormatHeaderKeepsCustomTitle ensures non-host titles remain visible.
func TestFormatHeaderKeepsCustomTitle(t *testing.T) {
	t.Parallel()

	session := tmux.Session{Name: "s", Windows: []tmux.Window{{Name: "win"}}}
	window := session.Windows[0]
	pane := tmux.Pane{
		Title:        "npm run dev",
		LastActivity: time.Now().Add(-time.Minute),
	}

	got := formatHeader(80, session, window, pane, false, false, false, false, "[x]", "dev-host")
	if !strings.Contains(got, "npm run dev") {
		t.Fatalf("formatHeader should keep custom title, got %q", got)
	}
}
