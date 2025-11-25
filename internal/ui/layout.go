// File layout.go contains helpers for sizing viewports and mapping mouse
// coordinates to cards.
package ui

import (
	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

type innerDimension struct {
	width  int
	height int
}

// viewportFor builds a viewport with sane defaults for capturing pane output.
func viewportFor(dim innerDimension) viewport.Model {
	opts := []viewport.Option{}
	if dim.width > 0 {
		opts = append(opts, viewport.WithWidth(dim.width))
	}
	height := minPreviewHeight
	if dim.height > 0 {
		height = dim.height
	}
	opts = append(opts, viewport.WithHeight(height))

	vp := viewport.New(opts...)
	vp.MouseWheelEnabled = false
	vp.MouseWheelDelta = scrollStep
	return vp
}

// updatePreviewDimensions recalculates viewport sizes based on terminal
// geometry and how many sessions are visible.
func (m *Model) updatePreviewDimensions(count int) {
	if count <= 0 || m.width <= 0 || m.height <= 0 {
		return
	}
	offset := m.previewOffset
	if offset <= 0 || offset >= m.height {
		offset = topPaddingLines
	}
	footerHeight := max(1, m.footerHeight)
	availableHeight := m.height - offset - footerHeight - gridSpacing
	if availableHeight < 1 {
		availableHeight = 1
	}
	const (
		frameHeight     = 3
		minInnerWidth   = 20
		columnOverhead  = cardPadding*2 + 2
		minColumnStride = minInnerWidth + columnOverhead
	)

	maxCols := 1
	if m.viewMode != viewModeDetail && count > 1 {
		maxCols = min(count, max(1, m.width/minColumnStride))
	}

	selectedCols := 1
	selectedHeight := 0
	selectedWidth := max(1, m.width-columnOverhead)

	for cols := maxCols; cols >= 1; cols-- {
		columnWidth := m.width / cols
		if columnWidth < minColumnStride && cols > 1 {
			continue
		}
		innerWidth := columnWidth - columnOverhead
		if innerWidth < 1 {
			innerWidth = 1
		}
		rows := (count + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
		bodyBudget := availableHeight - rows*frameHeight
		if bodyBudget < 0 {
			bodyBudget = 0
		}
		cellBody := 0
		if rows > 0 {
			cellBody = bodyBudget / rows
		}
		candidateHeight := 0
		switch {
		case cellBody > 2:
			candidateHeight = cellBody - 2
		case cellBody > 0:
			candidateHeight = 1
		default:
			candidateHeight = 0
		}

		predicted := rows * (frameHeight + candidateHeight)
		if cols == maxCols {
			selectedCols = cols
			selectedWidth = innerWidth
			selectedHeight = candidateHeight
		}
		if predicted <= availableHeight {
			// distribute leftover height evenly if possible
			slack := availableHeight - predicted
			if slack > 0 && rows > 0 {
				candidateHeight += slack / rows
				if candidateHeight < 0 {
					candidateHeight = 0
				}
			}
			selectedCols = cols
			selectedWidth = innerWidth
			selectedHeight = candidateHeight
			break
		}
	}

	if selectedWidth < 1 {
		selectedWidth = 1
	}
	if selectedHeight < 0 {
		selectedHeight = 0
	}

	m.cardCols = selectedCols
	m.cardInnerWidth = selectedWidth
	m.cardInnerHeight = selectedHeight
	for _, preview := range m.previews {
		if preview.viewport != nil {
			preview.viewport.SetWidth(m.cardInnerWidth)
			preview.viewport.SetHeight(m.cardInnerHeight)
		}
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

// ensureCursor keeps the cursor pointing at a visible session entry.
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

// moveCursorLeft shifts the cursor one column to the left when possible.
func (m *Model) moveCursorLeft() bool {
	return m.moveCursorByDelta(-1, true)
}

// moveCursorRight shifts the cursor one column to the right when possible.
func (m *Model) moveCursorRight() bool {
	return m.moveCursorByDelta(1, true)
}

// moveCursorUp moves the cursor up one row in the card grid.
func (m *Model) moveCursorUp() bool {
	cols := max(1, m.cardCols)
	return m.moveCursorByDelta(-cols, false)
}

// moveCursorDown moves the cursor down one row in the card grid.
func (m *Model) moveCursorDown() bool {
	cols := max(1, m.cardCols)
	return m.moveCursorByDelta(cols, false)
}

// moveCursorByDelta advances the cursor by the provided delta if permitted.
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
