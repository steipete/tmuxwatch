package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

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
		MarginTop(1).
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
		stale := m.isStale(session.ID)
		focused := session.ID == m.focusedSession
		headerContent := formatHeader(innerWidth, session, window, pane, focused, pulsing, stale)
		header := lipgloss.NewStyle().Render(headerContent)
		body := preview.viewport.View()

		borderStyle := baseStyle
		switch {
		case pane.Dead && pane.DeadStatus != 0:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitFail))
		case pane.Dead:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorExitOK))
		case stale:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorStale))
		case focused:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorFocus))
		case pulsing:
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(borderColorPulse))
		}

		card := borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body))

		rowIdx := idx / cols
		for len(grid) <= rowIdx {
			grid = append(grid, []string{})
		}
		grid[rowIdx] = append(grid[rowIdx], card)

		width := lipgloss.Width(card)
		height := lipgloss.Height(card)
		marginTop := borderStyle.GetMarginTop()
		marginLeft := borderStyle.GetMarginLeft()
		marginBottom := borderStyle.GetMarginBottom()
		marginRight := borderStyle.GetMarginRight()

		effectiveWidth := width - marginLeft - marginRight
		if effectiveWidth <= 0 {
			effectiveWidth = width
		}
		effectiveHeight := height - marginTop - marginBottom
		if effectiveHeight <= 0 {
			effectiveHeight = height
		}

		closeWidth := len(closeLabel)
		colIndex := idx % cols
		left := colIndex * cellWidth
		cardLeft := left + marginLeft
		cardWidth := effectiveWidth
		cardHeight := effectiveHeight
		closeRight := cardLeft + cardWidth - 1 - cardPadding
		if closeRight < cardLeft {
			closeRight = cardLeft
		}
		closeLeft := max(cardLeft, closeRight-closeWidth+1)
		bounds := cardBounds{
			sessionID:    session.ID,
			left:         cardLeft,
			width:        cardWidth,
			closeLeft:    closeLeft,
			closeRight:   closeRight,
			height:       cardHeight,
			marginTop:    marginTop,
			marginBottom: marginBottom,
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
			card.top = rowTop + card.marginTop
			if card.height <= 0 {
				card.height = rowHeight - card.marginTop - card.marginBottom
				if card.height <= 0 {
					card.height = rowHeight
				}
			}
			layoutIdx++
		}
		rendered = append(rendered, rowStr)
		cursorY += rowHeight
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// formatHeader builds the label line for a session card, colouring it based on
// status and focus state.
func formatHeader(width int, session tmux.Session, window tmux.Window, pane tmux.Pane, focused, pulsing, stale bool) string {
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
	case stale:
		style = style.Foreground(lipgloss.Color(headerColorStale))
	case focused:
		style = style.Foreground(lipgloss.Color(headerColorFocus))
	case pulsing:
		style = style.Foreground(lipgloss.Color(headerColorPulse))
	default:
		style = style.Foreground(lipgloss.Color(headerColorBase))
	}
	return style.Render(header)
}
