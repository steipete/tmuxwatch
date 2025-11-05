// File palette.go owns the lightweight command palette used to trigger common actions.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// openCommandPalette populates and displays the palette.
func (m *Model) openCommandPalette() {
	m.paletteCommands = m.buildCommandItems()
	m.paletteIndex = 0
	m.paletteOpen = true
}

// closePalette dismisses the palette without executing a command.
func (m *Model) closePalette() {
	m.paletteOpen = false
	m.paletteCommands = nil
	m.paletteIndex = 0
}

// handlePaletteKey processes keyboard input while the palette is open.
func (m *Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return m, nil
	}
	switch msg.String() {
	case "esc", "ctrl+p":
		m.closePalette()
		return m, nil
	case "up", "k":
		if len(m.paletteCommands) == 0 {
			return m, nil
		}
		m.paletteIndex--
		if m.paletteIndex < 0 {
			m.paletteIndex = len(m.paletteCommands) - 1
		}
		return m, nil
	case "down", "j":
		if len(m.paletteCommands) == 0 {
			return m, nil
		}
		m.paletteIndex = (m.paletteIndex + 1) % len(m.paletteCommands)
		return m, nil
	case "enter":
		if len(m.paletteCommands) == 0 {
			m.closePalette()
			return m, nil
		}
		item := m.paletteCommands[m.paletteIndex]
		m.closePalette()
		if !item.enabled || item.run == nil {
			return m, nil
		}
		return m, item.run(m)
	}
	return m, nil
}

// buildCommandItems constructs the available palette actions based on state.
func (m *Model) buildCommandItems() []commandItem {
	var items []commandItem

	items = append(items, m.tabPaletteCommands()...)

	focusedStale := m.isStale(m.focusedSession)
	if m.focusedSession != "" {
		sessionID := m.focusedSession
		items = append(items, commandItem{
			label:   "Kill focused stale session",
			enabled: focusedStale,
			run: func(*Model) tea.Cmd {
				if !m.isStale(sessionID) {
					return nil
				}
				return killSessionsCmd(m.client, []string{sessionID})
			},
		})
	}

	staleIDs := m.staleSessionIDs()
	items = append(items, commandItem{
		label:   fmt.Sprintf("Kill all stale sessions (%d)", len(staleIDs)),
		enabled: len(staleIDs) > 0,
		run: func(*Model) tea.Cmd {
			ids := m.staleSessionIDs()
			if len(ids) == 0 {
				return nil
			}
			return killSessionsCmd(m.client, ids)
		},
	})

	items = append(items, commandItem{
		label:   "Show hidden sessions",
		enabled: len(m.hidden) > 0,
		run: func(*Model) tea.Cmd {
			if len(m.hidden) == 0 {
				return nil
			}
			m.hidden = make(map[string]struct{})
			m.updatePreviewDimensions(m.filteredSessionCount())
			return nil
		},
	})

	items = append(items, commandItem{
		label:   "Force refresh from tmux",
		enabled: true,
		run: func(*Model) tea.Cmd {
			m.inflight = true
			return fetchSnapshotCmd(m.client)
		},
	})

	items = append(items, commandItem{
		label:   "Print card layout (stderr)",
		enabled: len(m.cardLayout) > 0,
		run: func(*Model) tea.Cmd {
			m.logCardLayout()
			return nil
		},
	})

	items = append(items, commandItem{
		label:   "Focus search bar (/)",
		enabled: !m.searching,
		run: func(*Model) tea.Cmd {
			if m.searching {
				return nil
			}
			m.searching = true
			m.searchInput.SetValue(m.searchQuery)
			m.searchInput.CursorEnd()
			return nil
		},
	})

	return items
}

// renderCommandPalette draws the palette overlay content with selection state.
func (m *Model) renderCommandPalette() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("231")).
		Render("command palette")

	if len(m.paletteCommands) == 0 {
		body := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Render("no actions available")
		return paletteStyle().Render(lipgloss.JoinVertical(lipgloss.Left, title, body))
	}

	var lines []string
	for i, item := range m.paletteCommands {
		marker := "  "
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
		if i == m.paletteIndex {
			marker = "▸ "
			labelStyle = labelStyle.Bold(true)
		}
		if !item.enabled {
			labelStyle = labelStyle.Foreground(lipgloss.Color("240"))
		}
		lines = append(lines, marker+labelStyle.Render(item.label))
	}

	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(lines, "\n"))

	return paletteStyle().Render(lipgloss.JoinVertical(lipgloss.Left, title, body))
}

func paletteStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		MarginTop(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Padding(1, 2)
}

// tabPaletteCommands returns palette items for switching between tabs.
func (m *Model) tabPaletteCommands() []commandItem {
	var cmds []commandItem

	if m.viewMode != viewModeOverview {
		cmds = append(cmds, commandItem{
			label:   "Switch to Overview tab",
			enabled: true,
			run: func(*Model) tea.Cmd {
				m.leaveDetail(false)
				m.setActiveTab(0)
				m.updatePreviewDimensions(m.filteredSessionCount())
				return nil
			},
		})
	}

	if m.detailSession != "" {
		label := fmt.Sprintf("Switch to Session tab (%s)", sessionLabel(m.detailSession))
		cmds = append(cmds, commandItem{
			label:   label,
			enabled: m.detailSession != "",
			run: func(*Model) tea.Cmd {
				if m.detailSession == "" {
					return nil
				}
				m.enterDetail(m.detailSession)
				m.updatePreviewDimensions(m.filteredSessionCount())
				return nil
			},
		})
	}

	return cmds
}
