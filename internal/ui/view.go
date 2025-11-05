// File view.go orchestrates the final Bubble Tea view composition.
package ui

import (
	"strings"

	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

const (
	topPaddingLines = 0
	gridSpacing     = 1
)

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

	separatorHeight := 0
	separator := ""
	if targetHeight > headerHeight+m.footerHeight {
		separatorHeight = 1
		separator = lipgloss.NewStyle().Width(targetWidth).Render("")
	}
	availableHeight := max(0, targetHeight-headerHeight-m.footerHeight-separatorHeight-gridSpacing)
	gridContent := m.renderSessionPreviews(headerHeight)
	if gridContent == "" {
		gridContent = emptyStateView(targetWidth, availableHeight)
	} else {
		gridContent = clampHeight(gridContent, availableHeight)
		gridContent = placeGridContent(gridContent, targetWidth, availableHeight)
	}

	footerView := status
	if m.footer != nil {
		m.footer.SetWidth(targetWidth)
		m.footer.SetHeight(m.footerHeight)
		m.footer.SetContent(status)
		footerView = m.footer.View()
	}

	filler := ""
	if gridSpacing > 0 && targetHeight > headerHeight+m.footerHeight+separatorHeight {
		filler = lipgloss.NewStyle().Width(targetWidth).Height(gridSpacing).Render("")
	}

	segments := []string{header, gridContent}
	if filler != "" {
		segments = append(segments, filler)
	}
	if separatorHeight > 0 {
		segments = append(segments, separator)
	}
	segments = append(segments, footerView)
	view := lipgloss.JoinVertical(lipgloss.Left, segments...)
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
	if limit <= 0 || content == "" {
		return ""
	}

	consumed := 0
	lines := 0
	for consumed < len(content) && lines < limit {
		remainder := content[consumed:]
		idx := strings.IndexByte(remainder, '\n')
		if idx == -1 {
			return content
		}
		consumed += idx + 1
		lines++
	}

	if lines < limit {
		return content
	}
	if consumed > 0 && content[consumed-1] == '\n' {
		consumed--
	}
	return content[:consumed]
}

func emptyStateView(width, height int) string {
	if width <= 0 {
		width = 40
	}
	message := "No tmux sessions detected."
	helper := "Start one with `tmux new -s demo`."
	box := lipgloss.JoinVertical(lipgloss.Left, message, helper)
	styled := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColorCursor)).
		Padding(1, 2).
		Foreground(lipgloss.Color("252")).
		Render(box)

	maxWidth := max(width, lipgloss.Width(styled))
	if height <= 0 {
		return lipgloss.PlaceHorizontal(maxWidth, lipgloss.Center, styled)
	}
	maxHeight := max(height, countLines(styled))
	return lipgloss.Place(maxWidth, maxHeight, lipgloss.Center, lipgloss.Center, styled)
}

func placeGridContent(content string, width, height int) string {
	if height <= 0 {
		return ""
	}
	if width <= 0 {
		width = 1
	}
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content, lipgloss.WithWhitespaceChars(" "))
}
