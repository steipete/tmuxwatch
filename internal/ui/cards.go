// File cards.go renders session preview cards and their visual chrome.
package ui

import (
	"fmt"
	"strings"
	"time"

	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/lipgloss/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

func decorateControl(label string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(headerColorFocus)).
		Render(label)
}

// renderSessionPreviews lays out each visible session card with consistent
// sizing and mouse hit-test metadata.
func (m *Model) renderSessionPreviews(offset int) string {
	sessions := m.filteredSessions()
	m.cardLayout = m.cardLayout[:0]
	if len(sessions) == 0 {
		m.cursorSession = ""
		return ""
	}

	cols := 1
	if m.viewMode != viewModeDetail && len(sessions) > 1 && m.width >= 70 {
		cols = 2
	}
	m.cardCols = cols
	m.ensureCursor(sessions)
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

		preview.viewport.SetWidth(innerWidth)
		preview.viewport.SetHeight(innerHeight)

		pulsing := now.Sub(preview.lastChanged) < pulseDuration
		stale := m.isStale(session.ID)
		focused := session.ID == m.focusedSession
		cursor := session.ID == m.cursorSession
		hovered := session.ID == m.hoveredSession

		cardID := fmt.Sprintf("%scard:%s", m.zonePrefix, session.ID)
		closeID := fmt.Sprintf("%sclose:%s", m.zonePrefix, session.ID)
		maxID := fmt.Sprintf("%smax:%s", m.zonePrefix, session.ID)
		collapseID := fmt.Sprintf("%scollapse:%s", m.zonePrefix, session.ID)

		maxLabel := maximizeLabel
		if m.viewMode == viewModeDetail && m.detailSession == session.ID {
			maxLabel = restoreLabel
		}
		collapseDisplay := collapseLabel
		if m.isCollapsed(session.ID) {
			collapseDisplay = expandLabel
		}

		maxContent := maxLabel
		collapseContent := collapseDisplay
		closeContent := closeLabel
		if m.hoveredControl == maxID {
			maxContent = decorateControl(maxLabel)
		}
		if m.hoveredControl == collapseID {
			collapseContent = decorateControl(collapseDisplay)
		}
		if m.hoveredControl == closeID {
			closeContent = decorateControl(closeLabel)
		}
		controlSegments := []string{
			zone.Mark(maxID, maxContent),
			zone.Mark(collapseID, collapseContent),
			zone.Mark(closeID, closeContent),
		}
		controls := strings.Join(controlSegments, " ")

		header := lipgloss.NewStyle().Render(formatHeader(innerWidth, session, window, pane, focused, pulsing, stale, cursor, controls))
		body := preview.viewport.View()
		if m.isCollapsed(session.ID) {
			body = ""
		}

		borderStyle := baseStyle
		switch {
		case pane.Dead && pane.DeadStatus != 0:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitFail))
		case pane.Dead:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitOK))
		case focused:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorFocus))
		case cursor:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorCursor))
		case hovered:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorHover))
		case stale:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorStale))
		case pulsing:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorPulse))
		}

		cardContent := borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
		cardContent = zone.Mark(cardID, cardContent)

		rowIdx := idx / cols
		for len(grid) <= rowIdx {
			grid = append(grid, []string{})
		}
		grid[rowIdx] = append(grid[rowIdx], cardContent)

		m.cardLayout = append(m.cardLayout, cardBounds{
			sessionID:      session.ID,
			zoneID:         cardID,
			closeZoneID:    closeID,
			maximizeZoneID: maxID,
			collapseZoneID: collapseID,
		})
	}

	var rendered []string
	for _, row := range grid {
		if len(row) == 0 {
			continue
		}
		padded := make([]string, 0, len(row))
		for _, card := range row {
			padded = append(padded, lipgloss.NewStyle().Width(cellWidth).Render(card))
		}
		rendered = append(rendered, lipgloss.JoinHorizontal(lipgloss.Left, padded...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// formatHeader builds the label line for a session card, colouring it based on
// status and focus state.
func formatHeader(width int, session tmux.Session, window tmux.Window, pane tmux.Pane, focused, pulsing, stale, cursor bool, controls string) string {
	var meta []string
	if pane.Dead {
		meta = append(meta, pane.StatusString())
	}
	if !pane.LastActivity.IsZero() {
		meta = append(meta, fmt.Sprintf("last %s", coarseDuration(time.Since(pane.LastActivity))))
	}
	label := fmt.Sprintf("%s · %s · %s", session.Name, window.Name, pane.TitleOrCmd())
	if stale {
		meta = append(meta, "stale")
	}

	if len(meta) > 0 {
		label += " · " + strings.Join(meta, " · ")
	}

	labelWidth := lipgloss.Width(label)
	spaceForLabel := width - lipgloss.Width(controls)
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
	header := label + strings.Repeat(" ", padding) + controls
	style := lipgloss.NewStyle()
	switch {
	case pane.Dead && pane.DeadStatus != 0:
		style = style.Foreground(lipgloss.Color(headerColorExitFail))
	case pane.Dead:
		style = style.Foreground(lipgloss.Color(headerColorExitOK))
	case focused:
		style = style.Foreground(lipgloss.Color(headerColorFocus))
	case cursor:
		style = style.Foreground(lipgloss.Color(headerColorCursor))
	case stale:
		style = style.Foreground(lipgloss.Color(headerColorStale))
	case pulsing:
		style = style.Foreground(lipgloss.Color(headerColorPulse))
	default:
		style = style.Foreground(lipgloss.Color(headerColorBase))
	}
	return style.Render(header)
}
