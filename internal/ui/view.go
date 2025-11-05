// File view.go orchestrates the final Bubble Tea view composition.
package ui

import (
	zone "github.com/alexanderbh/bubblezone/v2"
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
	var sections []string
	// Use a single-cell padding row so we can reuse it when spacing between sections.
	padding := lipgloss.NewStyle().Width(targetWidth).Render(" ")
	for i := 0; i < topPaddingLines; i++ {
		sections = append(sections, padding)
	}
	offset := topPaddingLines

	title := renderTitleBar(m)
	sections = append(sections, title)
	offset += lipgloss.Height(title)
	sections = append(sections, padding)
	offset++
	if m.searching {
		search := renderSearchBar(m.searchInput)
		sections = append(sections, search)
		offset += lipgloss.Height(search)
	} else if m.searchQuery != "" {
		summary := renderSearchSummary(m.searchQuery)
		sections = append(sections, summary)
		offset += lipgloss.Height(summary)
	}

	m.setActiveTab(m.activeTab)
	tabBar := m.renderTabBar(targetWidth)
	if tabBar != "" {
		sections = append(sections, tabBar)
		offset += lipgloss.Height(tabBar)
	}

	sections = append(sections, padding)
	offset++

	// previewOffset tells the mouse hit-test logic how many rows precede the grid.
	m.previewOffset = offset
	previews := m.renderSessionPreviews(offset)
	if previews == "" {
		sections = append(sections, lipgloss.NewStyle().Padding(1, 2).Render("No sessions to display."))
	} else {
		sections = append(sections, previews)
	}

	sections = append(sections, m.renderStatus())
	sections = append(sections, padding)
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)
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
