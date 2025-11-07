// File status.go renders the status footer, stale indicators, and toast
// messages.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// renderStatus returns the cached status footer, recomputing when necessary.
func (m *Model) renderStatus() string {
	content := m.buildStatusLine(m.width)
	if content == m.cachedStatus {
		return m.cachedStatus
	}
	m.cachedStatus = content
	return content
}

// buildStatusLine assembles the footer lines detailing input helpers, stale
// sessions, pane variables, toasts, and errors.
func (m *Model) buildStatusLine(width int) string {
	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 2).
			Render(fmt.Sprintf("mouse: click focus, scroll, %s/%s detail, %s/%s collapse, close %s · keys: / search, H show hidden, X kill stale, ctrl+X clean all, ctrl+P palette, q quit", maximizeLabel, restoreLabel, collapseLabel, expandLabel, closeLabel)),
	}

	if stale := m.staleSessionNames(); len(stale) > 0 {
		staleLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Padding(0, 2).
			Render(formatStaleLine(stale, width))
		lines = append(lines, staleLine)
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
	if toast := m.toastView(m.width); toast != "" {
		lines = append(lines, toast)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatStaleLine(names []string, width int) string {
	prefix := "stale sessions: "
	suffix := " (focus + X to clean)"
	if len(names) == 0 {
		return prefix + suffix
	}

	maxWidth := width
	if maxWidth <= 0 {
		maxWidth = lipgloss.Width(prefix) + lipgloss.Width(suffix) + 80
	}
	budget := maxWidth - lipgloss.Width(prefix) - lipgloss.Width(suffix)
	if budget < 0 {
		budget = 0
	}

	displayed := make([]string, 0, len(names))
	remaining := 0
	for i, name := range names {
		candidate := strings.Join(append(displayed, name), ", ")
		if lipgloss.Width(candidate) > budget && len(displayed) > 0 {
			remaining = len(names) - i
			break
		}
		displayed = append(displayed, name)
	}

	body := strings.Join(displayed, ", ")
	if remaining > 0 {
		if body != "" {
			body += " …"
		} else {
			body = "…"
		}
		body += fmt.Sprintf(" (+%d more)", remaining)
	}

	line := prefix + body + suffix
	// Final guard in case nothing fit.
	if body == "" {
		line = prefix + suffix
	}
	return line
}
