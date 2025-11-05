// File view.go orchestrates the final Bubble Tea view composition.
package ui

import (
	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"strings"
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

	status := m.renderStatus()
	m.footerHeight = max(1, countLines(status))
	m.updatePreviewDimensions(m.filteredSessionCount())

	gridLimit := max(1, targetHeight-headerHeight-m.footerHeight)
	gridContent := clampHeight(m.renderSessionPreviews(headerHeight), gridLimit)
	if gridContent == "" {
		gridContent = lipgloss.NewStyle().Padding(1, 2).Render("No sessions to display.")
	}

	footerView := status
	if m.footer != nil {
		m.footer.SetWidth(targetWidth)
		m.footer.SetHeight(m.footerHeight)
		m.footer.SetContent(status)
		footerView = m.footer.View()
	}

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		gridContent,
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

func clampHeight(content string, limit int) string {
	if limit <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= limit {
		return content
	}
	return strings.Join(lines[:limit], "\n")
}
