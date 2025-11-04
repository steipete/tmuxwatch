package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func renderSearchBar(input textinput.Model) string {
	label := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("62")).Render("Search")
	return lipgloss.JoinHorizontal(lipgloss.Left, label, input.View())
}

func renderSearchSummary(query string) string {
	return lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("62")).
		Render(fmt.Sprintf("Filter: %s (press / to edit, esc to clear)", query))
}

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
