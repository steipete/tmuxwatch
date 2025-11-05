// File state.go manages tab state, detail navigation, and collapse toggles.
package ui

import (
	"fmt"

	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/lipgloss/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// tabTitles derives the current tab titles based on overview and detail state.
var (
	tabActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(headerColorBase)).
		Background(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)

	tabInactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(headerColorCursor)).
		Padding(0, 1).
		MarginRight(1)
)

func (m *Model) tabTitles() []string {
	sessions := m.filteredSessionsFull()
	titles := make([]string, 1, len(sessions)+1)
	titles[0] = "Overview"
	m.tabSessionIDs = m.tabSessionIDs[:0]
	for _, session := range sessions {
		label := session.Name
		if label == "" {
			label = sessionLabel(session.ID)
		}
		m.tabSessionIDs = append(m.tabSessionIDs, session.ID)
		titles = append(titles, label)
	}
	return titles
}

// renderTabBar produces the tab strip view based on the active selection.
func (m *Model) renderTabBar(width int) string {
	titles := m.tabTitles()
	m.applyActiveTabSelection(len(titles))
	if len(titles) == 0 {
		return ""
	}
	segments := make([]string, 0, len(titles))
	for i, title := range titles {
		style := tabInactiveStyle
		if i == m.activeTab {
			style = tabActiveStyle
		}
		zoneID := fmt.Sprintf("%s###tab:%d", m.zonePrefix, i)
		segments = append(segments, zone.Mark(zoneID, style.Render(title)))
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Left, segments...)
	return lipgloss.PlaceHorizontal(max(width, lipgloss.Width(bar)), lipgloss.Left, bar)
}

// shiftActiveTab moves the active tab by the provided delta, clamping to range.
func (m *Model) shiftActiveTab(delta int) {
	m.setActiveTab(m.activeTab + delta)
}

// setActiveTab activates the provided tab index and adjusts the view mode.
func (m *Model) setActiveTab(idx int) {
	m.activeTab = idx
	titles := m.tabTitles()
	m.applyActiveTabSelection(len(titles))
}

// enterDetail switches the UI into detail mode for the given session.
func (m *Model) enterDetail(sessionID string) {
	if sessionID == "" {
		return
	}
	if _, ok := m.sessionByID(sessionID); !ok {
		return
	}
	m.detailSession = sessionID
	m.focusedSession = sessionID
	m.cursorSession = sessionID
	titles := m.tabTitles()
	idx := m.indexForSession(sessionID)
	if idx == 0 {
		idx = 1
	}
	m.activeTab = idx
	m.applyActiveTabSelection(len(titles))
}

// leaveDetail returns to overview mode and optionally clears detail metadata.
func (m *Model) leaveDetail(clear bool) {
	m.viewMode = viewModeOverview
	m.activeTab = 0
	if clear {
		m.detailSession = ""
	}
}

// handleDetailToggle toggles the detail view for the selected session.
func (m *Model) handleDetailToggle(sessionID string) {
	if sessionID == "" {
		return
	}
	if m.viewMode == viewModeDetail && m.detailSession == sessionID {
		m.leaveDetail(true)
		m.tabTitles()
		m.applyActiveTabSelection(len(m.tabSessionIDs) + 1)
		return
	}
	m.enterDetail(sessionID)
}

func (m *Model) indexForSession(id string) int {
	for i, sessionID := range m.tabSessionIDs {
		if sessionID == id {
			return i + 1
		}
	}
	return 0
}

func (m *Model) applyActiveTabSelection(length int) {
	if length <= 0 {
		m.activeTab = 0
		m.viewMode = viewModeOverview
		m.detailSession = ""
		return
	}
	if m.activeTab < 0 {
		m.activeTab = 0
	}
	if m.activeTab >= length {
		m.activeTab = length - 1
	}
	if m.activeTab == 0 || len(m.tabSessionIDs) == 0 {
		m.viewMode = viewModeOverview
		m.detailSession = ""
		return
	}
	index := m.activeTab - 1
	if index >= len(m.tabSessionIDs) {
		index = len(m.tabSessionIDs) - 1
		m.activeTab = index + 1
	}
	if index < 0 {
		m.activeTab = 0
		m.viewMode = viewModeOverview
		m.detailSession = ""
		return
	}
	sessionID := m.tabSessionIDs[index]
	m.detailSession = sessionID
	m.viewMode = viewModeDetail
	if m.focusedSession == "" {
		m.focusedSession = sessionID
	}
	if m.cursorSession == "" {
		m.cursorSession = sessionID
	}
}

// sessionByID locates a session by identifier if it exists in the snapshot.
func (m *Model) sessionByID(id string) (tmux.Session, bool) {
	for _, session := range m.sessions {
		if session.ID == id {
			return session, true
		}
	}
	return tmux.Session{}, false
}

// sessionExists reports whether the snapshot currently contains the session.
func (m *Model) sessionExists(id string) bool {
	_, ok := m.sessionByID(id)
	return ok
}

// isCollapsed reports whether the session card body is collapsed.
func (m *Model) isCollapsed(id string) bool {
	_, ok := m.collapsed[id]
	return ok
}

// toggleCollapsed flips the collapse state for the supplied session.
func (m *Model) toggleCollapsed(id string) {
	if id == "" {
		return
	}
	if m.isCollapsed(id) {
		delete(m.collapsed, id)
		return
	}
	m.collapsed[id] = struct{}{}
}

// clearCollapsed expands all sessions by clearing the collapsed set.
func (m *Model) clearCollapsed() {
	m.collapsed = make(map[string]struct{})
}
