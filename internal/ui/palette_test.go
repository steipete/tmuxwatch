// File palette_test.go exercises command palette keyboard interactions.
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestHandlePaletteKeyNavigationWrap ensures palette selection wraps via arrow keys.
func TestHandlePaletteKeyNavigationWrap(t *testing.T) {
	t.Parallel()

	m := &Model{paletteOpen: true}
	m.paletteCommands = []commandItem{{label: "first", enabled: true}, {label: "second", enabled: true}}

	_, cmd := m.handlePaletteKey(tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd != nil {
		t.Fatalf("down arrow should not return a command, got %v", cmd)
	}
	if m.paletteIndex != 1 {
		t.Fatalf("paletteIndex = %d, want 1", m.paletteIndex)
	}

	_, _ = m.handlePaletteKey(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.paletteIndex != 0 {
		t.Fatalf("paletteIndex should wrap to 0, got %d", m.paletteIndex)
	}

	_, _ = m.handlePaletteKey(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.paletteIndex != 1 {
		t.Fatalf("paletteIndex after up wrap = %d, want 1", m.paletteIndex)
	}
}

// TestHandlePaletteKeyExecute verifies enabled items execute and close the palette.
func TestHandlePaletteKeyExecute(t *testing.T) {
	t.Parallel()

	ran := false
	m := &Model{paletteOpen: true}
	m.paletteCommands = []commandItem{{
		label:   "run",
		enabled: true,
		run: func(*Model) tea.Cmd {
			return func() tea.Msg {
				ran = true
				return statusMsg("ok")
			}
		},
	}}

	model, cmd := m.handlePaletteKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model != m {
		t.Fatalf("handlePaletteKey returned unexpected model: %T", model)
	}
	if m.paletteOpen {
		t.Fatalf("paletteOpen should be false after execute")
	}
	if cmd == nil {
		t.Fatal("expected command to be returned for enabled item")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected status message, got %v", msg)
	}
	if !ran {
		t.Fatal("expected run func to execute")
	}
}

// TestTabPaletteCommands exposes overview/detail switching in the palette.
func TestTabPaletteCommands(t *testing.T) {
	t.Parallel()

	m := &Model{
		collapsed:     make(map[string]struct{}),
		hidden:        make(map[string]struct{}),
		stale:         make(map[string]struct{}),
		sessions:      []tmux.Session{{ID: "$1", Name: "dev"}},
		viewMode:      viewModeDetail,
		detailSession: "$1",
	}

	items := m.buildCommandItems()
	var hasOverview, hasDetail bool
	for _, item := range items {
		switch item.label {
		case "Switch to Overview tab":
			hasOverview = true
		case "Switch to Session tab (1)":
			hasDetail = true
		}
	}

	if !hasOverview {
		t.Fatal("expected palette to include Switch to Overview tab")
	}
	if !hasDetail {
		t.Fatal("expected palette to include Session tab command")
	}
}
