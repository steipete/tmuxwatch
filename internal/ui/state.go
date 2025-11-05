// File state.go manages tab state, detail navigation, and collapse toggles.
package ui

import (
	"fmt"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// tabTitles derives the current tab titles based on overview and detail state.
func (m *Model) tabTitles() []string {
	titles := []string{"Overview"}
	if m.detailSession != "" {
		if _, ok := m.sessionByID(m.detailSession); ok {
			titles = append(titles, fmt.Sprintf("Session %s", sessionLabel(m.detailSession)))
		} else {
			m.detailSession = ""
			m.viewMode = viewModeOverview
			m.activeTab = 0
		}
	}
	return titles
}

// renderTabBar produces the tab strip view based on the active selection.
func (m *Model) renderTabBar() string {
	titles := m.tabTitles()
	if len(titles) == 0 {
		return ""
	}
	if m.activeTab >= len(titles) {
		m.activeTab = len(titles) - 1
	}
	if m.activeTab < 0 {
		m.activeTab = 0
	}
	return m.tabRenderer.Render(titles, m.activeTab)
}

// shiftActiveTab moves the active tab by the provided delta, clamping to range.
func (m *Model) shiftActiveTab(delta int) {
	m.setActiveTab(m.activeTab + delta)
}

// setActiveTab activates the provided tab index and adjusts the view mode.
func (m *Model) setActiveTab(idx int) {
	titles := m.tabTitles()
	if len(titles) == 0 {
		m.activeTab = 0
		return
	}
	if idx < 0 {
		idx = 0
	} else if idx >= len(titles) {
		idx = len(titles) - 1
	}
	m.activeTab = idx
	if m.activeTab == 0 {
		m.viewMode = viewModeOverview
	} else if m.activeTab == 1 && m.detailSession != "" {
		m.viewMode = viewModeDetail
	}
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
	m.viewMode = viewModeDetail
	m.activeTab = 1
	m.focusedSession = sessionID
	m.cursorSession = sessionID
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
		return
	}
	m.enterDetail(sessionID)
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
