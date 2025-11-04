// File layout.go contains helpers for sizing viewports and mapping mouse
// coordinates to cards.
package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
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
