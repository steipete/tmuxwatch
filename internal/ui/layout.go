// File layout.go contains helpers for sizing viewports and mapping mouse
// coordinates to cards.
package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

type innerDimension struct {
	width  int
	height int
}

// viewportFor builds a viewport with sane defaults for capturing pane output.
func viewportFor(dim innerDimension) viewport.Model {
	vp := viewport.New(0, minPreviewHeight)
	vp.MouseWheelEnabled = false
	vp.MouseWheelDelta = scrollStep
	if dim.width > 0 {
		vp.Width = dim.width
	}
	if dim.height > 0 {
		vp.Height = dim.height
	}
	return vp
}

// updatePreviewDimensions recalculates viewport sizes based on terminal
// geometry and how many sessions are visible.
func (m *Model) updatePreviewDimensions(count int) {
	if count <= 0 || m.width <= 0 || m.height <= 0 {
		return
	}
	frameHeight := 3
	internalHeight := (m.height / count) - frameHeight
	if internalHeight < minPreviewHeight {
		internalHeight = minPreviewHeight
	}
	innerWidth := max(20, m.width-(cardPadding*2+2))
	for _, preview := range m.previews {
		preview.viewport.Width = innerWidth
		preview.viewport.Height = internalHeight
	}
}

// cardAt resolves the card located at the given mouse coordinates.
func (m *Model) cardAt(msg tea.MouseMsg) (cardBounds, bool) {
	for _, card := range m.cardLayout {
		if info := zone.Get(card.zoneID); info != nil && info.InBounds(msg) {
			return card, true
		}
	}
	return cardBounds{}, false
}

func (m *Model) ensureCursor(sessions []tmux.Session) {
	if len(sessions) == 0 {
		m.cursorSession = ""
		return
	}
	if m.cursorSession == "" {
		m.cursorSession = sessions[0].ID
		return
	}
	for _, session := range sessions {
		if session.ID == m.cursorSession {
			return
		}
	}
	m.cursorSession = sessions[0].ID
}

func (m *Model) moveCursorLeft() bool {
	return m.moveCursorByDelta(-1, true)
}

func (m *Model) moveCursorRight() bool {
	return m.moveCursorByDelta(1, true)
}

func (m *Model) moveCursorUp() bool {
	cols := max(1, m.cardCols)
	return m.moveCursorByDelta(-cols, false)
}

func (m *Model) moveCursorDown() bool {
	cols := max(1, m.cardCols)
	return m.moveCursorByDelta(cols, false)
}

func (m *Model) moveCursorByDelta(delta int, enforceRow bool) bool {
	sessions := m.filteredSessions()
	if len(sessions) == 0 {
		m.cursorSession = ""
		return false
	}
	m.ensureCursor(sessions)
	cols := max(1, m.cardCols)
	currentIndex := -1
	for idx, session := range sessions {
		if session.ID == m.cursorSession {
			currentIndex = idx
			break
		}
	}
	if currentIndex == -1 {
		return false
	}
	nextIndex := currentIndex + delta
	if nextIndex < 0 || nextIndex >= len(sessions) {
		return false
	}
	if enforceRow {
		currentRow := currentIndex / cols
		nextRow := nextIndex / cols
		if currentRow != nextRow {
			return false
		}
	}
	m.cursorSession = sessions[nextIndex].ID
	return true
}
