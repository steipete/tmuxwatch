// File render.go renders the Bubble Tea view and supporting components used by
// tmuxwatch.
package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// View renders the entire tmuxwatch interface, including title bar, search
// state, session previews, and status footer.
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
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderSessionPreviews lays out each visible session card with consistent
// sizing and mouse hit-test metadata.
func (m *Model) renderSessionPreviews(offset int) string {
	sessions := m.filteredSessions()
	m.cardLayout = m.cardLayout[:0]
	if len(sessions) == 0 {
		return ""
	}

	cols := 1
	if len(sessions) > 1 && m.width >= 70 {
		cols = 2
	}
	baseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColorBase)).
		Padding(0, cardPadding)

	innerWidth := max(20, (m.width/cols)-(cardPadding*2+2))
	cellWidth := innerWidth + cardPadding*2 + 2
	rows := (len(sessions) + cols - 1) / cols
	innerHeight := max(minPreviewHeight, (m.height-offset-3)/max(1, rows))

	grid := make([][]string, 0)
	now := time.Now()

	for idx, session := range sessions {
		window, ok := activeWindow(session)
		if !ok {
			continue
		}
		pane, ok := activePane(window)
		if !ok {
			continue
		}
		preview, ok := m.previews[session.ID]
		if !ok {
			continue
		}

		preview.viewport.Width = innerWidth
		preview.viewport.Height = innerHeight

		pulsing := now.Sub(preview.lastChanged) < pulseDuration
		focused := session.ID == m.focusedSession
		headerContent := formatHeader(innerWidth, session, window, pane, focused, pulsing)
		header := lipgloss.NewStyle().Render(headerContent)
		body := preview.viewport.View()

		borderStyle := baseStyle
		switch {
		case pane.Dead && pane.DeadStatus != 0:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitFail))
		case pane.Dead:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitOK))
		case focused:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorFocus))
		case pulsing:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorPulse))
		}

		card := borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body))

		rowIdx := idx / cols
		if len(grid) <= rowIdx {
			grid = append(grid, []string{})
		}
		grid[rowIdx] = append(grid[rowIdx], card)

		width := lipgloss.Width(card)
		height := lipgloss.Height(card)
		closeWidth := len(closeLabel)
		colIndex := idx % cols
		left := colIndex * cellWidth
		right := left + cellWidth - 1
		contentRight := left + width - 1
		if contentRight < left {
			contentRight = left
		}
		closeRight := contentRight - 1 - cardPadding
		if closeRight > contentRight {
			closeRight = contentRight
		}
		if closeRight < left {
			closeRight = left
		}
		closeLeft := max(left, closeRight-closeWidth+1)
		bounds := cardBounds{
			sessionID:  session.ID,
			left:       left,
			right:      right,
			closeLeft:  closeLeft,
			closeRight: closeRight,
			height:     height,
		}
		m.cardLayout = append(m.cardLayout, bounds)
	}

	var rendered []string
	layoutIdx := 0
	cursorY := 0
	for _, row := range grid {
		if len(row) == 0 {
			continue
		}
		padded := make([]string, 0, len(row))
		for _, card := range row {
			padded = append(padded, lipgloss.NewStyle().Width(cellWidth).Render(card))
		}
		rowStr := lipgloss.JoinHorizontal(lipgloss.Left, padded...)
		rowHeight := lipgloss.Height(rowStr)
		if rowHeight <= 0 {
			rowHeight = 1
		}
		rowTop := offset + cursorY
		for range row {
			if layoutIdx >= len(m.cardLayout) {
				break
			}
			card := &m.cardLayout[layoutIdx]
			card.top = rowTop
			h := card.height
			if h <= 0 {
				h = rowHeight
			}
			card.bottom = rowTop + h - 1
			layoutIdx++
		}
		rendered = append(rendered, rowStr)
		cursorY += rowHeight
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// formatHeader builds the label line for a session card, colouring it based on
// status and focus state.
func formatHeader(width int, session tmux.Session, window tmux.Window, pane tmux.Pane, focused, pulsing bool) string {
	var meta []string
	if pane.Dead {
		meta = append(meta, pane.StatusString())
	}
	if !pane.LastActivity.IsZero() {
		meta = append(meta, fmt.Sprintf("last %s", coarseDuration(time.Since(pane.LastActivity))))
	}
	label := fmt.Sprintf("%s · %s · %s", session.Name, window.Name, pane.TitleOrCmd())
	if len(meta) > 0 {
		label += " · " + strings.Join(meta, " · ")
	}
	labelWidth := lipgloss.Width(label)
	spaceForLabel := width - len(closeLabel)
	if spaceForLabel < 1 {
		spaceForLabel = 1
	}
	if labelWidth > spaceForLabel {
		label = lipgloss.NewStyle().Width(spaceForLabel).MaxWidth(spaceForLabel).Render(label)
	}
	padding := spaceForLabel - lipgloss.Width(label)
	if padding < 0 {
		padding = 0
	}
	header := label + strings.Repeat(" ", padding) + closeLabel
	style := lipgloss.NewStyle()
	switch {
	case pane.Dead && pane.DeadStatus != 0:
		style = style.Foreground(lipgloss.Color(headerColorExitFail))
	case pane.Dead:
		style = style.Foreground(lipgloss.Color(headerColorExitOK))
	case focused:
		style = style.Foreground(lipgloss.Color(headerColorFocus))
	case pulsing:
		style = style.Foreground(lipgloss.Color(headerColorPulse))
	default:
		style = style.Foreground(lipgloss.Color(headerColorBase))
	}
	return style.Render(header)
}

// renderStatus builds or reuses the status line shown at the bottom of the UI.
func (m *Model) renderStatus() string {
	content := m.buildStatusLine()
	if content == m.cachedStatus {
		return m.cachedStatus
	}
	m.cachedStatus = content
	return content
}

// buildStatusLine highlights controls and light status details at the bottom of
// the screen.
func (m *Model) buildStatusLine() string {
	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(1, 2).
			Render("mouse: click focus, scroll logs, close [x] · keys: / search, H show hidden, q quit"),
	}

	if preview, ok := m.previews[m.focusedSession]; ok && len(preview.vars) > 0 {
		varsLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 2).
			Render(formatPaneVariables(preview.vars))
		lines = append(lines, varsLine)
	}

	if m.err != nil {
		errPart := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Padding(0, 2).
			Render("Error: " + m.err.Error())
		lines = append(lines, errPart)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderSearchBar formats the interactive search prompt.
func renderSearchBar(input textinput.Model) string {
	label := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("62")).Render("Search")
	return lipgloss.JoinHorizontal(lipgloss.Left, label, input.View())
}

// renderSearchSummary shows the active search filter when the prompt is closed.
func renderSearchSummary(query string) string {
	return lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("62")).
		Render(fmt.Sprintf("Filter: %s (press / to edit, esc to clear)", query))
}

// renderTitleBar prints the tmuxwatch heading.
func renderTitleBar(m *Model) string {
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("62")).
		Padding(0, 2)
	name := baseStyle.Copy().Bold(true).Render("tmuxwatch")

	metaParts := []string{fmt.Sprintf("%d sessions", len(m.sessions))}
	if !m.lastUpdated.IsZero() {
		metaParts = append(metaParts, fmt.Sprintf("refreshed %s ago", coarseDuration(time.Since(m.lastUpdated))))
	}
	if m.focusedSession != "" {
		metaParts = append(metaParts, "focus "+m.focusedSession)
	}
	if m.searchQuery != "" {
		metaParts = append(metaParts, fmt.Sprintf("filter %q", m.searchQuery))
	}
	if len(metaParts) == 0 {
		return name
	}
	meta := baseStyle.Copy().
		Foreground(lipgloss.Color("249")).
		Render(strings.Join(metaParts, " • "))
	return lipgloss.JoinHorizontal(lipgloss.Left, name, meta)
}

func formatPaneVariables(vars map[string]string) string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, vars[k]))
	}
	return "vars: " + strings.Join(parts, " ")
}
