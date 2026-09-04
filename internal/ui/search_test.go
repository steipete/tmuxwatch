package ui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
	"github.com/steipete/tmuxwatch/internal/zone"
)

func TestSearchAcceptsTypingAndReopens(t *testing.T) {
	zone.NewGlobal()
	for _, entry := range []string{"slash", "ctrl+f", "palette"} {
		t.Run(entry, func(t *testing.T) {
			m := NewModel(nil, time.Second, nil, false)
			m.sessions = []tmux.Session{{ID: "$1", Name: "alpha"}, {ID: "$2", Name: "beta"}}
			_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
			open := func() {
				switch entry {
				case "slash":
					_, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
				case "ctrl+f":
					_, _ = m.Update(tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})
				case "palette":
					_, _ = m.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
					found := false
					for i, item := range m.paletteCommands {
						if item.label == "Focus search bar (/)" {
							m.paletteIndex = i
							found = true
							break
						}
					}
					if !found {
						t.Fatal("search palette command missing")
					}
					_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				}
			}
			open()
			for _, r := range "alpha" {
				_, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
			}
			if m.searchQuery != "alpha" || m.filteredSessionCount() != 1 {
				t.Fatalf("typing should filter to alpha: query=%q, sessions=%d", m.searchQuery, m.filteredSessionCount())
			}
			_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if m.searching || m.searchInput.Focused() {
				t.Fatal("enter should close and blur search")
			}
			open()
			_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
			if m.searchQuery != "alph" {
				t.Fatalf("reopened search should remain editable, got %q", m.searchQuery)
			}
			_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			if m.searching || m.searchInput.Focused() {
				t.Fatal("escape should close and blur search")
			}
		})
	}
}
