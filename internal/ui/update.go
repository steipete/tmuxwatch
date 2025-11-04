// File update.go contains the Bubble Tea update loop and helper workflow that
// react to incoming messages.
package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// Update processes Bubble Tea messages and routes them to specialised
// handlers, returning the next command to execute.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePreviewDimensions(m.filteredSessionCount())
	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchKey(msg)
		}
		if handled, cmd := m.handleGlobalKey(msg); handled {
			return m, cmd
		}
		if handled, cmd := m.handleFocusedKey(msg); handled {
			return m, cmd
		}
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case searchBlurMsg:
		m.searching = false
    case snapshotMsg:
        m.inflight = false
        m.err = nil
        m.lastUpdated = msg.snapshot.Timestamp
        m.sessions = msg.snapshot.Sessions
        m.updateStaleSessions()
        cmd := m.ensurePreviewsAndCapture()
        m.updatePreviewDimensions(m.filteredSessionCount())
        return m, tea.Batch(scheduleTick(m.pollInterval), cmd)
    case errMsg:
        m.inflight = false
        m.err = msg.err
        return m, scheduleTick(m.pollInterval)
    case paneContentMsg:
		if preview, ok := m.previews[msg.sessionID]; ok && preview.paneID == msg.paneID {
			content := strings.TrimRight(msg.text, "\n")
			if msg.err != nil {
				content = "Pane capture error: " + msg.err.Error()
			}
			if content != preview.lastContent {
				preview.viewport.SetContent(content)
				preview.lastContent = content
				preview.lastChanged = time.Now()
				preview.viewport.GotoBottom()
			}
		}
    case paneVarsMsg:
        if preview, ok := m.previews[msg.sessionID]; ok && preview.paneID == msg.paneID {
            if msg.err != nil {
                preview.vars = map[string]string{"error": msg.err.Error()}
            } else {
                preview.vars = msg.vars
            }
        }
    case killSessionMsg:
        m.inflight = true
        if m.focusedSession == msg.sessionID {
            m.focusedSession = ""
        }
        delete(m.previews, msg.sessionID)
        delete(m.hidden, msg.sessionID)
        delete(m.stale, msg.sessionID)
        return m, fetchSnapshotCmd(m.client)
    case tickMsg:
        if m.inflight {
            return m, nil
        }
        m.inflight = true
		return m, fetchSnapshotCmd(m.client)
	}
	return m, nil
}

// ensurePreviewsAndCapture keeps track of per-session previews and captures
// fresh content for their active panes.
func (m *Model) ensurePreviewsAndCapture() tea.Cmd {
	active := make(map[string]struct{}, len(m.sessions))
	var cmds []tea.Cmd
	for _, session := range m.sessions {
		if m.isHidden(session.ID) {
			continue
		}
		active[session.ID] = struct{}{}
		window, ok := activeWindow(session)
		if !ok {
			continue
		}
		pane, ok := activePane(window)
		if !ok {
			continue
		}
		preview := m.previews[session.ID]
		if preview == nil {
			vp := viewportFor(innerDimension{
				width:  m.width,
				height: m.height,
			})
			preview = &sessionPreview{viewport: &vp, lastChanged: time.Now()}
			m.previews[session.ID] = preview
		}
        if preview.paneID != pane.ID {
            preview.viewport.SetContent("")
            preview.paneID = pane.ID
            preview.lastContent = ""
            preview.vars = nil
        }
        cmds = append(cmds, fetchPaneContentCmd(m.client, session.ID, pane.ID, 400))
        if session.ID == m.focusedSession {
            cmds = append(cmds, fetchPaneVarsCmd(m.client, session.ID, pane.ID))
        }
	}
	for sessionID := range m.previews {
		if _, ok := active[sessionID]; !ok {
			delete(m.previews, sessionID)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// filteredSessions applies the active search filter and hidden toggles to the
// current snapshot.
func (m *Model) filteredSessions() []tmux.Session {
	var out []tmux.Session
	query := strings.ToLower(m.searchQuery)
	for _, session := range m.sessions {
		if m.isHidden(session.ID) {
			continue
		}
		if query == "" || sessionMatches(session, query) {
			out = append(out, session)
		}
	}
	return out
}

// filteredSessionCount provides a quick count for layout calculations.
func (m *Model) filteredSessionCount() int {
	return len(m.filteredSessions())
}
