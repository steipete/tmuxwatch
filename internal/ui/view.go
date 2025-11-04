package ui

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// View renders the entire tmuxwatch interface, including title bar, search
// state, session previews, status footer, and overlays.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	var sections []string
	padding := lipgloss.NewStyle().Width(max(m.width, 1)).Render(" ")
	sections = append(sections, padding, padding, padding, padding)
	offset := 4

	title := renderTitleBar(m)
	sections = append(sections, title)
	offset += lipgloss.Height(title)
	if m.searching {
		search := renderSearchBar(m.searchInput)
		sections = append(sections, search)
		offset += lipgloss.Height(search)
	} else if m.searchQuery != "" {
		summary := renderSearchSummary(m.searchQuery)
		sections = append(sections, summary)
		offset += lipgloss.Height(summary)
	}

	sections = append(sections, padding)
	offset++

	previews := m.renderSessionPreviews(offset)
	if previews == "" {
		sections = append(sections, lipgloss.NewStyle().Padding(1, 2).Render("No sessions to display."))
	} else {
		sections = append(sections, previews)
	}

	sections = append(sections, m.renderStatus())
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

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
