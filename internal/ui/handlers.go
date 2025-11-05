// File handlers.go groups input handling routines for keyboard and mouse
// events.
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	zone "github.com/alexanderbh/bubblezone/v2"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// handleGlobalKey processes keys that apply regardless of focus.
func (m *Model) handleGlobalKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return false, nil
	}
	if msg.String() != "esc" {
		m.lastEsc = time.Time{}
	}
	switch msg.String() {
	case "shift+left":
		m.shiftActiveTab(-1)
		m.updatePreviewDimensions(m.filteredSessionCount())
		return true, nil
	case "shift+right":
		m.shiftActiveTab(1)
		m.updatePreviewDimensions(m.filteredSessionCount())
		return true, nil
	case "left":
		if m.focusedSession != "" {
			return false, nil
		}
		if m.moveCursorLeft() {
			return true, nil
		}
		return true, nil
	case "right":
		if m.focusedSession != "" {
			return false, nil
		}
		if m.moveCursorRight() {
			return true, nil
		}
		return true, nil
	case "up":
		if m.focusedSession != "" {
			return false, nil
		}
		if m.moveCursorUp() {
			return true, nil
		}
		return true, nil
	case "down":
		if m.focusedSession != "" {
			return false, nil
		}
		if m.moveCursorDown() {
			return true, nil
		}
		return true, nil
	case "enter":
		if m.cursorSession == "" {
			return true, nil
		}
		if m.focusedSession != m.cursorSession {
			m.focusedSession = m.cursorSession
			m.resetCtrlC()
			if preview, ok := m.previews[m.focusedSession]; ok {
				preview.viewport.GotoBottom()
				if preview.paneID != "" {
					return true, fetchPaneVarsCmd(m.client, m.focusedSession, preview.paneID)
				}
			}
		}
		return true, nil
	case "/", "ctrl+f":
		m.resetCtrlC()
		m.searching = true
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.CursorEnd()
		return true, nil
	case "esc":
		if m.viewMode == viewModeDetail && m.activeTab == 1 {
			m.leaveDetail(false)
			m.updatePreviewDimensions(m.filteredSessionCount())
			return true, nil
		}
		if m.searchQuery != "" {
			m.resetCtrlC()
			m.searchQuery = ""
			m.updatePreviewDimensions(m.filteredSessionCount())
			return true, nil
		}
		if m.focusedSession == "" {
			return false, nil
		}
		now := time.Now()
		if !m.lastEsc.IsZero() && now.Sub(m.lastEsc) < quitChordWindow {
			previous := m.focusedSession
			m.focusedSession = ""
			m.cursorSession = previous
			m.lastEsc = time.Time{}
			m.resetCtrlC()
			return true, nil
		}
		m.lastEsc = now
		return true, nil
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
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return false, nil
	}
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
	case "ctrl+m":
		target := m.focusedSession
		if target == "" {
			target = m.cursorSession
		}
		if target != "" {
			m.handleDetailToggle(target)
			m.updatePreviewDimensions(m.filteredSessionCount())
		}
		return true, nil
	case "z":
		if m.focusedSession != "" {
			m.toggleCollapsed(m.focusedSession)
			m.updatePreviewDimensions(m.filteredSessionCount())
			return true, nil
		}
	case "Z":
		m.clearCollapsed()
		m.updatePreviewDimensions(m.filteredSessionCount())
		return true, nil
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
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}
	key := press.Key()
	switch key.Code {
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
	}
	if key.Mod&tea.ModAlt != 0 {
		return nil, false
	}
	if key.Text == "" {
		return nil, false
	}
	return []string{key.Text}, true
}

// handleMouse wires up focus toggles, pane hiding, and scroll gestures.
func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if handled, cmd := m.handleTabMouse(msg); handled {
		return m, cmd
	}
	if len(m.cardLayout) == 0 {
		return m, nil
	}
	card, ok := m.cardAt(msg)
	m.logMouseEvent(msg, card, ok)
	if !ok {
		return m, nil
	}
	preview := m.previews[card.sessionID]
	mouse := msg.Mouse()
	switch mouse.Button {
	case tea.MouseWheelDown:
		if _, wheel := msg.(tea.MouseWheelMsg); wheel && preview != nil {
			preview.viewport.ScrollDown(scrollStep)
		}
	case tea.MouseWheelUp:
		if _, wheel := msg.(tea.MouseWheelMsg); wheel && preview != nil {
			preview.viewport.ScrollUp(scrollStep)
		}
	case tea.MouseLeft:
		if _, click := msg.(tea.MouseClickMsg); click {
			if info := zone.Get(card.maximizeZoneID); info != nil && info.InBounds(msg) {
				m.handleDetailToggle(card.sessionID)
				m.updatePreviewDimensions(m.filteredSessionCount())
				return m, nil
			}
			if info := zone.Get(card.collapseZoneID); info != nil && info.InBounds(msg) {
				m.toggleCollapsed(card.sessionID)
				m.updatePreviewDimensions(m.filteredSessionCount())
				return m, nil
			}
			if info := zone.Get(card.closeZoneID); info != nil && info.InBounds(msg) {
				m.hidden[card.sessionID] = struct{}{}
				if m.focusedSession == card.sessionID {
					m.focusedSession = ""
				}
				if m.cursorSession == card.sessionID {
					m.cursorSession = ""
				}
				delete(m.previews, card.sessionID)
				m.resetCtrlC()
				m.updatePreviewDimensions(m.filteredSessionCount())
				return m, showStatusMessage(fmt.Sprintf("Closed session %s", sessionLabel(card.sessionID)))
			}
			m.focusedSession = card.sessionID
			m.cursorSession = card.sessionID
			m.resetCtrlC()
			if preview != nil {
				preview.viewport.GotoBottom()
				if preview.paneID != "" {
					return m, fetchPaneVarsCmd(m.client, card.sessionID, preview.paneID)
				}
			}
		}
	}
	return m, nil
}

// handleTabMouse reacts to mouse input overlapping the tab strip and switches
// tabs when the user clicks the corresponding title.
func (m *Model) handleTabMouse(msg tea.MouseMsg) (bool, tea.Cmd) {
	mouse := msg.Mouse()
	if mouse.Button != tea.MouseLeft {
		return false, nil
	}
	switch msg.(type) {
	case tea.MouseWheelMsg, tea.MouseMotionMsg:
		return false, nil
	}
	manager := zone.DefaultManager
	if manager == nil {
		return false, nil
	}
	ids := manager.IDsInBounds(msg)
	if len(ids) == 0 {
		return false, nil
	}
	tabIndex, ok := tabIndexFromZoneIDs(ids)
	if !ok {
		return false, nil
	}
	switch msg.(type) {
	case tea.MouseClickMsg, tea.MouseReleaseMsg:
		m.setActiveTab(tabIndex)
		m.updatePreviewDimensions(m.filteredSessionCount())
		return true, nil
	}
	return false, nil
}

// tabIndexFromZoneIDs inspects zone identifiers and extracts the tab index
// encoded by BubbleApp's `tabtitles` component.
func tabIndexFromZoneIDs(ids []string) (int, bool) {
	for _, rawID := range ids {
		child := rawID
		if idx := strings.LastIndex(rawID, "###"); idx >= 0 {
			child = rawID[idx+3:]
		}
		if !strings.HasPrefix(child, "tab:") {
			continue
		}
		parts := strings.SplitN(child, ":", 2)
		if len(parts) != 2 {
			continue
		}
		tabIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		return tabIndex, true
	}
	return 0, false
}

// handleSearchKey updates the search field when the user is actively editing.
func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return m, nil
	}
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
