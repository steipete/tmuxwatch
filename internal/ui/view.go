package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the entire tmuxwatch interface, including title bar, search
// state, session previews, status footer, and overlays.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	var sections []string
	title := renderTitleBar(m)
	sections = append(sections, title)
	offset := lipgloss.Height(title)
	if m.searching {
		search := renderSearchBar(m.searchInput)
		sections = append(sections, search)
		offset += lipgloss.Height(search)
	} else if m.searchQuery != "" {
		summary := renderSearchSummary(m.searchQuery)
		sections = append(sections, summary)
		offset += lipgloss.Height(summary)
	}

	previews := m.renderSessionPreviews(offset)
	if previews == "" {
		sections = append(sections, lipgloss.NewStyle().Padding(1, 2).Render("No sessions to display."))
	} else {
		sections = append(sections, previews)
	}

	sections = append(sections, m.renderStatus())
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	if !m.paletteOpen {
		return view
	}

	palette := m.renderCommandPalette()
	width := max(m.width, lipgloss.Width(view))
	height := max(m.height, countLines(view))

	overlay := lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		palette,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("235")),
	)

	return overlayView(view, overlay, width, height)
}
