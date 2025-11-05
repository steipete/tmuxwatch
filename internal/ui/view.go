// File view.go orchestrates the final Bubble Tea view composition.
package ui

import (
	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/bubbles/v2/viewport"
	"github.com/charmbracelet/lipgloss/v2"
)

const topPaddingLines = 0

// View renders the entire tmuxwatch interface, including title bar, search
// state, session previews, status footer, and overlays.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	targetWidth := max(m.width, 1)
	targetHeight := max(m.height, 1)
	padding := lipgloss.NewStyle().Width(targetWidth).Render(" ")

	headerParts := []string{renderTitleBar(m)}
	if m.searching {
		headerParts = append(headerParts, renderSearchBar(m.searchInput))
	} else if m.searchQuery != "" {
		headerParts = append(headerParts, renderSearchSummary(m.searchQuery))
	}

	m.setActiveTab(m.activeTab)
	if tabBar := m.renderTabBar(targetWidth); tabBar != "" {
		headerParts = append(headerParts, tabBar)
	}
	headerParts = append(headerParts, padding)

	header := lipgloss.JoinVertical(lipgloss.Left, headerParts...)
	headerHeight := max(1, countLines(header))
	m.previewOffset = headerHeight

	gridContent := m.renderSessionPreviews(headerHeight)
	if gridContent == "" {
		gridContent = lipgloss.NewStyle().Padding(1, 2).Render("No sessions to display.")
	}

	status := m.renderStatus()
	cardHeight := max(1, targetHeight-headerHeight-max(1, m.footerHeight+1))
	grid := viewport.New(viewport.WithHeight(cardHeight), viewport.WithWidth(targetWidth))
	grid.MouseWheelEnabled = false
	grid.SetContent(gridContent)

	var footerView string
	if m.footer != nil {
		m.footer.SetWidth(targetWidth)
		height := max(1, countLines(status))
		m.footer.SetHeight(height)
		m.footer.SetContent(status)
		m.footerHeight = height
		footerView = m.footer.View()
	} else {
		m.footerHeight = max(1, countLines(status))
		footerView = status
	}

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		grid.View(),
		footerView,
	)
	view = lipgloss.Place(targetWidth, targetHeight, lipgloss.Left, lipgloss.Top, view)

	if m.traceMouse {
		m.logCardLayout()
	}

	if !m.paletteOpen {
		return zone.Scan(view)
	}

	palette := m.renderCommandPalette()
	paletteWidth := lipgloss.Width(palette)
	paletteHeight := countLines(palette)
	width := max(m.width, max(lipgloss.Width(view), paletteWidth))
	height := max(m.height, max(countLines(view), paletteHeight))
	offsetX := max((width-paletteWidth)/2, 0)
	offsetY := max((height-paletteHeight)/2, 0)

	return zone.Scan(overlayView(view, palette, width, height, offsetX, offsetY))
}
