// File chrome.go defines UI chrome helpers such as search bars and headers.
package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/textinput"
	"github.com/charmbracelet/lipgloss/v2"
)

// renderSearchBar prints the interactive search prompt and input box.
func renderSearchBar(input textinput.Model) string {
	label := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("62")).Render("Search")
	return lipgloss.JoinHorizontal(lipgloss.Left, label, input.View())
}

// renderSearchSummary shows the current filter query when the search box is
// closed.
func renderSearchSummary(query string) string {
	return lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("62")).
		Render(fmt.Sprintf("Filter: %s (press / to edit, esc to clear)", query))
}

// renderTitleBar constructs the application header with live metadata.
func renderTitleBar(m *Model, width int) string {
	width = max(width, 1)

	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("62"))
	name := base.Copy().Bold(true).Padding(0, 2).Render("tmuxwatch")

	totalSessions := len(m.sessions)
	staleCount := len(m.stale)
	activeCount := totalSessions - staleCount
	if activeCount < 0 {
		activeCount = 0
	}

	summary := fmt.Sprintf("%d sessions (%d active, %d stale)", totalSessions, activeCount, staleCount)
	metaParts := []string{summary}
	if !m.lastUpdated.IsZero() {
		metaParts = append(metaParts, fmt.Sprintf("refreshed %s ago", coarseDuration(time.Since(m.lastUpdated))))
	}
	if m.focusedSession != "" {
		metaParts = append(metaParts, "focus "+m.focusedSession)
	}
	if m.searchQuery != "" {
		metaParts = append(metaParts, fmt.Sprintf("filter %q", m.searchQuery))
	}
	content := name
	if len(metaParts) > 0 {
		meta := base.Copy().
			Padding(0, 2).
			Foreground(lipgloss.Color("249")).
			Render(strings.Join(metaParts, " • "))
		content = lipgloss.JoinHorizontal(lipgloss.Left, content, meta)
	}

	remaining := width - lipgloss.Width(content)
	if remaining > 0 {
		padding := base.Render(strings.Repeat(" ", remaining))
		content += padding
	}
	return content
}

// formatPaneVariables formats sorted tmux pane variables for display.
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
