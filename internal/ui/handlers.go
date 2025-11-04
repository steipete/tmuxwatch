// File handlers.go groups input handling routines for keyboard and mouse
// events.
package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// handleGlobalKey processes keys that apply regardless of focus.
func (m *Model) handleGlobalKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "/", "ctrl+f":
		m.resetCtrlC()
		m.searching = true
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.CursorEnd()
		return true, nil
	case "esc":
		if m.searchQuery != "" {
			m.resetCtrlC()
			m.searchQuery = ""
			m.updatePreviewDimensions(m.filteredSessionCount())
			return true, nil
		}
		return false, nil
	case "ctrl+p":
		if m.paletteOpen {
			m.closePalette()
		} else {
			m.openCommandPalette()
		}
		return true, nil
	case "H":
		if len(m.hidden) > 0 {
			m.resetCtrlC()
			m.hidden = make(map[string]struct{})
			m.updatePreviewDimensions(m.filteredSessionCount())
		}
		return true, nil
	case "ctrl+x":
		ids := m.staleSessionIDs()
		if len(ids) == 0 {
			return true, nil
		}
		m.resetCtrlC()
		return true, killSessionsCmd(m.client, ids)
	case "q":
		m.resetCtrlC()
		return true, tea.Quit
	case "X":
		if m.focusedSession == "" {
			return true, nil
		}
		if !m.isStale(m.focusedSession) {
			return true, nil
		}
		m.resetCtrlC()
		return true, killSessionsCmd(m.client, []string{m.focusedSession})
	}
	return false, nil
}

// handleFocusedKey forwards navigation and control keys to the focused pane or
// local viewport.
func (m *Model) handleFocusedKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.focusedSession == "" {
		return false, nil
	}
	preview, ok := m.previews[m.focusedSession]
	if !ok {
		return false, nil
	}
	pane, paneOK := m.paneFor(m.focusedSession)
	switch msg.String() {
	case "up":
		m.resetCtrlC()
		preview.viewport.ScrollUp(1)
		return true, nil
	case "down":
		m.resetCtrlC()
		preview.viewport.ScrollDown(1)
		return true, nil
	case "pgup":
		m.resetCtrlC()
		preview.viewport.PageUp()
		return true, nil
	case "pgdown":
		m.resetCtrlC()
		preview.viewport.PageDown()
		return true, nil
	case "ctrl+u":
		m.resetCtrlC()
		preview.viewport.ScrollUp(scrollStep)
		return true, nil
	case "ctrl+d":
		m.resetCtrlC()
		preview.viewport.ScrollDown(scrollStep)
		return true, nil
	case "g":
		m.resetCtrlC()
		preview.viewport.GotoTop()
		return true, nil
	case "G":
		m.resetCtrlC()
		preview.viewport.GotoBottom()
		return true, nil
	case "ctrl+c":
		now := time.Now()
		if !paneOK || pane.Dead || preview.paneID == "" {
			return true, tea.Quit
		}
		cmd := sendKeysCmd(m.client, preview.paneID, "C-c")
		if !m.lastCtrlC.IsZero() && now.Sub(m.lastCtrlC) < quitChordWindow {
			m.resetCtrlC()
			return true, tea.Batch(cmd, tea.Quit)
		}
		m.lastCtrlC = now
		return true, cmd
	}

	keys, ok := tmuxKeysFrom(msg)
	if !ok || preview.paneID == "" {
		m.resetCtrlC()
		return false, nil
	}
	m.resetCtrlC()
	return true, sendKeysCmd(m.client, preview.paneID, keys...)
}

// tmuxKeysFrom converts Bubble Tea key messages into tmux key strings.
func tmuxKeysFrom(msg tea.KeyMsg) ([]string, bool) {
	switch msg.Type {
	case tea.KeyEnter:
		return []string{"Enter"}, true
	case tea.KeyTab:
		return []string{"Tab"}, true
	case tea.KeySpace:
		return []string{" "}, true
	case tea.KeyBackspace:
		return []string{"BSpace"}, true
	case tea.KeyDelete:
		return []string{"Delete"}, true
	case tea.KeyEsc:
		return []string{"Escape"}, true
	case tea.KeyRunes:
		if msg.Alt {
			return nil, false
		}
		return []string{string(msg.Runes)}, true
	}
	return nil, false
}

// handleMouse wires up focus toggles, pane hiding, and scroll gestures.
func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
    if len(m.cardLayout) == 0 {
        return m, nil
    }
    card, ok := m.cardAt(msg)
    m.logMouseEvent(msg, card, ok)
    if !ok {
        return m, nil
    }
    preview := m.previews[card.sessionID]
	switch {
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		if preview != nil {
			_ = preview.viewport.ScrollDown(scrollStep)
		}
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		if preview != nil {
			_ = preview.viewport.ScrollUp(scrollStep)
		}
    case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
        if info := zone.Get(card.closeZoneID); info != nil && info.InBounds(msg) {
            m.hidden[card.sessionID] = struct{}{}
            if m.focusedSession == card.sessionID {
                m.focusedSession = ""
            }
            delete(m.previews, card.sessionID)
            m.resetCtrlC()
            m.updatePreviewDimensions(m.filteredSessionCount())
            return m, nil
        }
        m.focusedSession = card.sessionID
        m.resetCtrlC()
        if preview != nil {
            preview.viewport.GotoBottom()
            if preview.paneID != "" {
                return m, fetchPaneVarsCmd(m.client, card.sessionID, preview.paneID)
            }
        }
	}
	return m, nil
}

// handleSearchKey updates the search field when the user is actively editing.
func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
		m.searchInput.Blur()
		return m, nil
	case "enter":
		m.searchQuery = strings.TrimSpace(m.searchInput.Value())
		m.searching = false
		m.searchInput.Blur()
		m.updatePreviewDimensions(m.filteredSessionCount())
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = strings.TrimSpace(m.searchInput.Value())
	m.updatePreviewDimensions(m.filteredSessionCount())
	return m, cmd
}

// isHidden reports whether the given session ID is hidden from the grid.
func (m *Model) isHidden(id string) bool {
	_, ok := m.hidden[id]
	return ok
}

// resetCtrlC clears the timing cache used to detect the quit chord.
func (m *Model) resetCtrlC() {
	m.lastCtrlC = time.Time{}
}

// paneFor resolves the active pane for a session, returning false when the
// session has no panes.
func (m *Model) paneFor(sessionID string) (tmux.Pane, bool) {
	for _, session := range m.sessions {
		if session.ID == sessionID {
			window, ok := activeWindow(session)
			if !ok {
				return tmux.Pane{}, false
			}
			return activePane(window)
		}
	}
	return tmux.Pane{}, false
}
