package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

const (
	defaultPollInterval = time.Second
	sidePadding         = 2
	minPreviewHeight    = 5
)

type (
	snapshotMsg    struct{ snapshot tmux.Snapshot }
	paneContentMsg struct {
		sessionID, paneID, text string
		err                     error
	}
	errMsg        struct{ err error }
	tickMsg       struct{}
	searchBlurMsg struct{}
)

type sessionPreview struct {
	viewport *viewport.Model
	paneID   string
}

// Model drives the Bubble Tea UI.
type Model struct {
	client       *tmux.Client
	pollInterval time.Duration

	width  int
	height int

	sessions []tmux.Session

	previews map[string]*sessionPreview

	searchInput textinput.Model
	searching   bool
	searchQuery string

	lastUpdated time.Time
	err         error
	inflight    bool
}

// NewModel constructs a Model with sane defaults.
func NewModel(client *tmux.Client, poll time.Duration) Model {
	if poll <= 0 {
		poll = defaultPollInterval
	}
	ti := textinput.New()
	ti.Placeholder = "filter sessions, windows, panes"
	ti.CharLimit = 256
	ti.Width = 60
	return Model{
		client:       client,
		pollInterval: poll,
		previews:     make(map[string]*sessionPreview),
		searchInput:  ti,
		inflight:     true,
	}
}

// Init starts polling tmux.
func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchSnapshotCmd(m.client), scheduleTick(m.pollInterval))
}

// Update handles Bubble Tea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePreviewDimensions(m.filteredSessionCount())
	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchKey(msg)
		}
		switch msg.String() {
		case "/", "ctrl+f":
			m.searching = true
			m.searchInput.SetValue(m.searchQuery)
			m.searchInput.CursorEnd()
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case searchBlurMsg:
		m.searching = false
	case snapshotMsg:
		m.inflight = false
		m.err = nil
		m.lastUpdated = msg.snapshot.Timestamp
		m.sessions = msg.snapshot.Sessions
		cmd := m.ensurePreviewsAndCapture()
		m.updatePreviewDimensions(m.filteredSessionCount())
		return m, tea.Batch(scheduleTick(m.pollInterval), cmd)
	case errMsg:
		m.inflight = false
		m.err = msg.err
		return m, scheduleTick(m.pollInterval)
	case paneContentMsg:
		preview, ok := m.previews[msg.sessionID]
		if !ok || preview.paneID != msg.paneID {
			return m, nil
		}
		content := "Error capturing pane."
		if msg.err == nil {
			content = strings.TrimRight(msg.text, "\n")
		}
		preview.viewport.SetContent(content)
		preview.viewport.GotoBottom()
	case tickMsg:
		if m.inflight {
			return m, nil
		}
		m.inflight = true
		return m, fetchSnapshotCmd(m.client)
	}
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
		m.searchInput.Blur()
		return m, nil
	case "enter":
		m.searchQuery = strings.TrimSpace(m.searchInput.Value())
		m.searching = false
		m.searchInput.Blur()
		m.updatePreviewDimensions(m.filteredSessionCount())
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = strings.TrimSpace(m.searchInput.Value())
	m.updatePreviewDimensions(m.filteredSessionCount())
	return m, cmd
}

func (m *Model) ensurePreviewsAndCapture() tea.Cmd {
	active := make(map[string]struct{}, len(m.sessions))
	var cmds []tea.Cmd
	for _, session := range m.sessions {
		active[session.ID] = struct{}{}
		window, ok := activeWindow(session)
		if !ok {
			continue
		}
		pane, ok := activePane(window)
		if !ok {
			continue
		}
		preview, exists := m.previews[session.ID]
		if !exists {
			vp := viewport.New(0, minPreviewHeight)
			preview = &sessionPreview{viewport: &vp}
			m.previews[session.ID] = preview
		}
		if preview.paneID != pane.ID {
			preview.viewport.SetContent("")
			preview.paneID = pane.ID
		}
		cmds = append(cmds, fetchPaneContentCmd(m.client, session.ID, pane.ID, 400))
	}
	for sessionID := range m.previews {
		if _, ok := active[sessionID]; !ok {
			delete(m.previews, sessionID)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) filteredSessionCount() int {
	return len(m.filteredSessions())
}

func (m Model) filteredSessions() []tmux.Session {
	if m.searchQuery == "" {
		return m.sessions
	}
	var filtered []tmux.Session
	q := strings.ToLower(m.searchQuery)
	for _, session := range m.sessions {
		if sessionMatches(session, q) {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

// View renders the interface.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	var rows []string
	if m.searching {
		rows = append(rows, renderSearchBar(m.searchInput))
	} else if m.searchQuery != "" {
		rows = append(rows, renderSearchSummary(m.searchQuery))
	}

	previews := m.renderSessionPreviews()
	if previews == "" {
		rows = append(rows, lipgloss.NewStyle().Padding(1, 2).Render("No sessions match the current filter."))
	} else {
		rows = append(rows, previews)
	}
	rows = append(rows, m.renderStatus())
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderSessionPreviews() string {
	sessions := m.filteredSessions()
	if len(sessions) == 0 {
		return ""
	}

	var rendered []string
	for _, session := range sessions {
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
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Bold(true).
			Render(fmt.Sprintf("%s · %s · %s", session.Name, window.Name, pane.TitleOrCmd()))

		body := preview.viewport.View()
		card := lipgloss.NewStyle().
			Margin(0, 1).
			Padding(0, 1).
			BorderForeground(lipgloss.Color("62")).
			BorderStyle(lipgloss.RoundedBorder()).
			Width(preview.viewport.Width + sidePadding).
			Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
		rendered = append(rendered, card)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

func (m *Model) updatePreviewDimensions(count int) {
	if count <= 0 || m.width == 0 || m.height == 0 {
		return
	}
	usableHeight := m.height - 2 // leave room for status/search
	if usableHeight < count*minPreviewHeight {
		usableHeight = count * minPreviewHeight
	}
	heightPer := max(minPreviewHeight, usableHeight/count)
	width := max(10, m.width-2*sidePadding)
	for _, preview := range m.previews {
		preview.viewport.Width = width - 4
		preview.viewport.Height = heightPer - 3
		if preview.viewport.Height < 1 {
			preview.viewport.Height = 1
		}
	}
}

func (m Model) renderStatus() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 2)
	last := "never"
	if !m.lastUpdated.IsZero() {
		last = fmt.Sprintf("%s ago", humanizeDuration(time.Since(m.lastUpdated)))
	}
	errPart := ""
	if m.err != nil {
		errPart = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("Error: " + m.err.Error())
	}
	filterPart := ""
	if m.searchQuery != "" {
		filterPart = fmt.Sprintf(" | filter: %s", m.searchQuery)
	}
	info := fmt.Sprintf("sessions: %d | last refresh: %s%s", len(m.sessions), last, filterPart)
	help := "keys: / search · q quit"
	status := lipgloss.JoinHorizontal(lipgloss.Left, style.Render(info), lipgloss.NewStyle().PaddingLeft(2).Render(help))
	if errPart != "" {
		status = lipgloss.JoinHorizontal(lipgloss.Left, status, lipgloss.NewStyle().PaddingLeft(2).Render(errPart))
	}
	return status
}

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

func sessionMatches(session tmux.Session, query string) bool {
	if strings.Contains(strings.ToLower(session.Name), query) {
		return true
	}
	for _, window := range session.Windows {
		if strings.Contains(strings.ToLower(window.Name), query) {
			return true
		}
		for _, pane := range window.Panes {
			if strings.Contains(strings.ToLower(pane.TitleOrCmd()), query) {
				return true
			}
		}
	}
	return false
}

func activeWindow(session tmux.Session) (tmux.Window, bool) {
	for _, window := range session.Windows {
		if window.Active {
			return window, true
		}
	}
	if len(session.Windows) > 0 {
		return session.Windows[0], true
	}
	return tmux.Window{}, false
}

func activePane(window tmux.Window) (tmux.Pane, bool) {
	for _, pane := range window.Panes {
		if pane.Active {
			return pane, true
		}
	}
	if len(window.Panes) > 0 {
		return window.Panes[0], true
	}
	return tmux.Pane{}, false
}

func humanizeDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}

func fetchSnapshotCmd(client *tmux.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		snap, err := client.Snapshot(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return snapshotMsg{snapshot: snap}
	}
}

func scheduleTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func fetchPaneContentCmd(client *tmux.Client, sessionID, paneID string, lines int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		text, err := client.CapturePane(ctx, paneID, lines)
		return paneContentMsg{sessionID: sessionID, paneID: paneID, text: text, err: err}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
