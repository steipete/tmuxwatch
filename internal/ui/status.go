// File status.go renders the status footer, stale indicators, and toast
// messages.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderStatus returns the cached status footer, recomputing when necessary.
func (m *Model) renderStatus() string {
	content := m.buildStatusLine()
	if content == m.cachedStatus {
		return m.cachedStatus
	}
	m.cachedStatus = content
	return content
}

// buildStatusLine assembles the footer lines detailing input helpers, stale
// sessions, pane variables, toasts, and errors.
func (m *Model) buildStatusLine() string {
	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(1, 2).
			Render("mouse: click focus, scroll logs, close [x] · keys: / search, H show hidden, X kill stale, ctrl+X clean all, ctrl+P palette, q quit"),
	}

	if stale := m.staleSessionNames(); len(stale) > 0 {
		staleLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Padding(0, 2).
			Render("stale sessions: " + strings.Join(stale, ", ") + " (focus + X to clean)")
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
