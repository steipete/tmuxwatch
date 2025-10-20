package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusPanes
)

// Model encapsulates the Bubble Tea state machine for the tmuxwatch UI.
type Model struct {
	client       *tmux.Client
	pollInterval time.Duration

	width  int
	height int

	sessions []tmux.Session
	sidebar  []sidebarItem

	sidebarIndex int

	selectedWindowID string
	selectedPaneID   string

	hidden map[string]struct{}

	focus focusArea

	captureLines int

	paneContents map[string]paneContent
	paneOrder    map[string][]string

	lastUpdated time.Time
	err         error

	inflight bool
}

type paneContent struct {
	text    string
	err     error
	fetched time.Time
}

type sidebarItem struct {
	sessionIdx int
	windowIdx  int
	windowID   string
}

// snapshotMsg carries the tmux snapshot back into the Bubble Tea model.
type snapshotMsg struct {
	snapshot tmux.Snapshot
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

type tickMsg struct{}

type paneContentMsg struct {
	paneID string
	text   string
	err    error
}

// NewModel returns a ready-to-run Model.
func NewModel(client *tmux.Client, poll time.Duration) Model {
	if poll <= 0 {
		poll = time.Second
	}
	return Model{
		client:       client,
		pollInterval: poll,
		hidden:       make(map[string]struct{}),
		focus:        focusSidebar,
		sidebarIndex: -1,
		inflight:     true, // We'll fetch immediately in Init.
		captureLines: 200,
		paneContents: make(map[string]paneContent),
		paneOrder:    make(map[string][]string),
	}
}

// Init triggers the initial fetch of tmux state.
func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchSnapshotCmd(m.client), scheduleTick(m.pollInterval))
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	case snapshotMsg:
		m.inflight = false
		m.err = nil
		m.lastUpdated = msg.snapshot.Timestamp
		m.sessions = msg.snapshot.Sessions
		m.syncPaneOrder(msg.snapshot.Sessions)
		m.rebuildSidebar(msg.snapshot.Sessions)
		m.ensureSelection()
		cmds := []tea.Cmd{scheduleTick(m.pollInterval)}
		// Refresh pane content whenever snapshot updates.
		if m.selectedPaneID != "" {
			cmds = append(cmds, fetchPaneContentCmd(m.client, m.selectedPaneID, m.captureLines))
		}
		return m, tea.Batch(cmds...)
	case errMsg:
		m.inflight = false
		m.err = msg.err
		return m, scheduleTick(m.pollInterval)
	case paneContentMsg:
		if msg.err != nil {
			m.paneContents[msg.paneID] = paneContent{
				err:     msg.err,
				fetched: time.Now(),
			}
		} else {
			m.paneContents[msg.paneID] = paneContent{
				text:    msg.text,
				fetched: time.Now(),
			}
		}
		return m, nil
	case tickMsg:
		if m.inflight {
			return m, nil
		}
		m.inflight = true
		return m, fetchSnapshotCmd(m.client)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		if m.focus == focusSidebar && len(m.visiblePanes()) > 0 {
			m.focus = focusPanes
		} else {
			m.focus = focusSidebar
		}
	case "shift+tab":
		if m.focus == focusPanes {
			m.focus = focusSidebar
		}
	case "r":
		if m.inflight {
			return m, nil
		}
		m.inflight = true
		return m, fetchSnapshotCmd(m.client)
	}

	switch m.focus {
	case focusSidebar:
		return m.handleSidebarKey(msg)
	case focusPanes:
		return m.handlePaneKey(msg)
	}

	return m, nil
}

func (m Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.sidebar) == 0 {
		return m, nil
	}
	prevPane := m.selectedPaneID
	switch msg.String() {
	case "j", "down":
		m.sidebarIndex++
		if m.sidebarIndex >= len(m.sidebar) {
			m.sidebarIndex = len(m.sidebar) - 1
		}
		m.updateSelectedWindow()
	case "k", "up":
		m.sidebarIndex--
		if m.sidebarIndex < 0 {
			m.sidebarIndex = 0
		}
		m.updateSelectedWindow()
	case "enter", "right", "l":
		if len(m.visiblePanes()) > 0 {
			m.focus = focusPanes
		}
	}
	if prevPane != m.selectedPaneID && m.selectedPaneID != "" {
		return m, fetchPaneContentCmd(m.client, m.selectedPaneID, m.captureLines)
	}
	return m, nil
}

func (m Model) handlePaneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	panes := m.visiblePanes()
	if len(panes) == 0 {
		if msg.String() == "H" {
			m.hidden = make(map[string]struct{})
			refreshed := m.visiblePanes()
			if len(refreshed) > 0 {
				m.selectedPaneID = refreshed[0].ID
				return m, fetchPaneContentCmd(m.client, m.selectedPaneID, m.captureLines)
			}
		}
		return m, nil
	}
	prevSelection := m.selectedPaneID
	currentIndex := m.currentPaneIndex(panes)
	switch msg.String() {
	case "j", "down":
		if currentIndex < len(panes)-1 {
			currentIndex++
		}
		m.selectedPaneID = panes[currentIndex].ID
	case "k", "up":
		if currentIndex > 0 {
			currentIndex--
		}
		m.selectedPaneID = panes[currentIndex].ID
	case "right", "l":
		if currentIndex < len(panes)-1 {
			currentIndex++
			m.selectedPaneID = panes[currentIndex].ID
		}
	case "left":
		if currentIndex > 0 {
			currentIndex--
			m.selectedPaneID = panes[currentIndex].ID
		}
	case "h":
		if len(panes) == 0 {
			break
		}
		paneID := panes[currentIndex].ID
		m.hidden[paneID] = struct{}{}
		newPanes := m.visiblePanes()
		if len(newPanes) == 0 {
			m.selectedPaneID = ""
		} else if currentIndex >= len(newPanes) {
			m.selectedPaneID = newPanes[len(newPanes)-1].ID
		} else {
			m.selectedPaneID = newPanes[currentIndex].ID
		}
	case "H":
		m.hidden = make(map[string]struct{})
		panes := m.visiblePanes()
		if len(panes) > 0 {
			m.selectedPaneID = panes[0].ID
		} else {
			m.selectedPaneID = ""
		}
	case "[":
		m.reorderSelectedPane(-1)
	case "]":
		m.reorderSelectedPane(1)
	}
	var cmd tea.Cmd
	if prevSelection != m.selectedPaneID && m.selectedPaneID != "" {
		cmd = fetchPaneContentCmd(m.client, m.selectedPaneID, m.captureLines)
	}
	return m, cmd
}

func (m *Model) rebuildSidebar(sessions []tmux.Session) {
	m.sidebar = m.sidebar[:0]
	for si, session := range sessions {
		for wi, window := range session.Windows {
			m.sidebar = append(m.sidebar, sidebarItem{
				sessionIdx: si,
				windowIdx:  wi,
				windowID:   window.ID,
			})
		}
	}
}

func (m *Model) syncPaneOrder(sessions []tmux.Session) {
	seenWindows := make(map[string]struct{})
	seenPanes := make(map[string]struct{})

	for _, session := range sessions {
		for _, window := range session.Windows {
			seenWindows[window.ID] = struct{}{}

			currentOrder := append([]string(nil), m.paneOrder[window.ID]...)
			present := make(map[string]struct{}, len(window.Panes))
			for _, pane := range window.Panes {
				present[pane.ID] = struct{}{}
				seenPanes[pane.ID] = struct{}{}
			}

			filtered := filteredByPresent(currentOrder, present)
			existing := make(map[string]struct{}, len(filtered))
			for _, id := range filtered {
				existing[id] = struct{}{}
			}
			for _, pane := range window.Panes {
				if _, ok := existing[pane.ID]; !ok {
					filtered = append(filtered, pane.ID)
					existing[pane.ID] = struct{}{}
				}
			}
			if len(filtered) == 0 {
				for _, pane := range window.Panes {
					filtered = append(filtered, pane.ID)
				}
			}
			m.paneOrder[window.ID] = filtered
		}
	}

	for windowID := range m.paneOrder {
		if _, ok := seenWindows[windowID]; !ok {
			delete(m.paneOrder, windowID)
		}
	}
	for paneID := range m.paneContents {
		if _, ok := seenPanes[paneID]; !ok {
			delete(m.paneContents, paneID)
		}
	}
	for paneID := range m.hidden {
		if _, ok := seenPanes[paneID]; !ok {
			delete(m.hidden, paneID)
		}
	}
}

func filteredByPresent(order []string, present map[string]struct{}) []string {
	if len(order) == 0 {
		return order
	}
	out := make([]string, 0, len(order))
	for _, id := range order {
		if _, ok := present[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func (m *Model) ensureSelection() {
	if len(m.sidebar) == 0 {
		m.sidebarIndex = -1
		m.selectedWindowID = ""
		m.selectedPaneID = ""
		return
	}

	if m.selectedWindowID == "" {
		m.sidebarIndex = 0
		m.selectedWindowID = m.sidebar[0].windowID
	}

	// ensure sidebar index matches selectedWindowID
	found := false
	for idx, item := range m.sidebar {
		if item.windowID == m.selectedWindowID {
			m.sidebarIndex = idx
			found = true
			break
		}
	}
	if !found {
		m.sidebarIndex = 0
		m.selectedWindowID = m.sidebar[0].windowID
	}

	panes := m.visiblePanes()
	if len(panes) == 0 {
		m.selectedPaneID = ""
		return
	}
	if m.selectedPaneID == "" {
		m.selectedPaneID = panes[0].ID
		return
	}
	for _, pane := range panes {
		if pane.ID == m.selectedPaneID {
			return
		}
	}
	m.selectedPaneID = panes[0].ID
}

func (m *Model) updateSelectedWindow() {
	if len(m.sidebar) == 0 {
		m.selectedWindowID = ""
		m.selectedPaneID = ""
		return
	}
	if m.sidebarIndex < 0 {
		m.sidebarIndex = 0
	}
	if m.sidebarIndex >= len(m.sidebar) {
		m.sidebarIndex = len(m.sidebar) - 1
	}
	selected := m.sidebar[m.sidebarIndex]
	m.selectedWindowID = selected.windowID
	panes := m.visiblePanes()
	if len(panes) == 0 {
		m.selectedPaneID = ""
		return
	}
	m.selectedPaneID = panes[0].ID
}

func (m Model) visiblePanes() []tmux.Pane {
	window, ok := m.currentWindow()
	if !ok {
		return nil
	}
	if len(window.Panes) == 0 {
		return nil
	}
	order := m.paneOrder[window.ID]
	if len(order) == 0 {
		order = make([]string, 0, len(window.Panes))
		for _, pane := range window.Panes {
			order = append(order, pane.ID)
		}
	}
	paneLookup := make(map[string]tmux.Pane, len(window.Panes))
	for _, pane := range window.Panes {
		paneLookup[pane.ID] = pane
	}
	panes := make([]tmux.Pane, 0, len(order))
	for _, paneID := range order {
		pane, ok := paneLookup[paneID]
		if !ok {
			continue
		}
		if _, hidden := m.hidden[paneID]; hidden {
			continue
		}
		panes = append(panes, pane)
	}
	return panes
}

func (m Model) currentWindow() (tmux.Window, bool) {
	for _, item := range m.sidebar {
		if item.windowID == m.selectedWindowID {
			session := m.sessions[item.sessionIdx]
			if item.windowIdx < len(session.Windows) {
				return session.Windows[item.windowIdx], true
			}
			break
		}
	}
	return tmux.Window{}, false
}

func (m Model) currentPaneIndex(panes []tmux.Pane) int {
	if m.selectedPaneID == "" {
		return 0
	}
	for idx, pane := range panes {
		if pane.ID == m.selectedPaneID {
			return idx
		}
	}
	return 0
}

func (m *Model) reorderSelectedPane(delta int) {
	if m.selectedPaneID == "" {
		return
	}
	window, ok := m.currentWindow()
	if !ok {
		return
	}
	order := m.paneOrder[window.ID]
	if len(order) == 0 {
		for _, pane := range window.Panes {
			order = append(order, pane.ID)
		}
	}
	index := -1
	for i, id := range order {
		if id == m.selectedPaneID {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	target := index + delta
	if target < 0 || target >= len(order) {
		return
	}
	order[index], order[target] = order[target], order[index]
	m.paneOrder[window.ID] = order
}

// View renders the TUI.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	sidebar := m.renderSidebar()
	main := m.renderPaneView()
	status := m.renderStatus()
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func (m Model) renderSidebar() string {
	width := 30
	style := lipgloss.NewStyle().Width(width).Padding(0, 1)
	selectedStyle := style.Copy().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62"))
	activeStyle := style.Copy().Foreground(lipgloss.Color("14"))
	var builder strings.Builder
	if len(m.sidebar) == 0 {
		builder.WriteString(selectedStyle.Render("No sessions"))
		return builder.String()
	}
	for idx, item := range m.sidebar {
		session := m.sessions[item.sessionIdx]
		window := session.Windows[item.windowIdx]
		label := fmt.Sprintf("%s:%d %s", session.Name, window.Index, window.Name)
		if window.Active {
			label = activeStyle.Render(label)
		} else {
			label = style.Render(label)
		}
		if idx == m.sidebarIndex {
			if m.focus == focusSidebar {
				label = selectedStyle.Render(label)
			} else {
				label = selectedStyle.Copy().Background(lipgloss.Color("238")).Render(label)
			}
		}
		builder.WriteString(label)
		if idx < len(m.sidebar)-1 {
			builder.WriteRune('\n')
		}
	}
	return builder.String()
}

func (m Model) renderPaneView() string {
	window, ok := m.currentWindow()
	width := max(m.width-30, 0)
	style := lipgloss.NewStyle().Padding(0, 1).Width(width)
	if !ok {
		return style.Render("No window selected")
	}
	panes := m.visiblePanes()
	if len(window.Panes) > 0 && len(panes) == 0 {
		return style.Render("All panes hidden. Press H to show them.")
	}
	order := m.paneOrder[window.ID]
	tabStyle := lipgloss.NewStyle().Padding(0, 1).MarginRight(1)
	selectedTabStyle := tabStyle.Copy().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62"))
	inactiveTabStyle := tabStyle.Copy().Foreground(lipgloss.Color("244"))

	paneLookup := make(map[string]tmux.Pane, len(window.Panes))
	for _, pane := range window.Panes {
		paneLookup[pane.ID] = pane
	}

	var tabs []string
	for _, paneID := range order {
		pane, ok := paneLookup[paneID]
		if !ok {
			continue
		}
		if _, hidden := m.hidden[paneID]; hidden {
			continue
		}
		label := pane.TitleOrCmd()
		if paneID == m.selectedPaneID && m.focus == focusPanes {
			tabs = append(tabs, selectedTabStyle.Render(label))
		} else if paneID == m.selectedPaneID {
			tabs = append(tabs, selectedTabStyle.Copy().Background(lipgloss.Color("238")).Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}

	if len(tabs) == 0 {
		return style.Render("No panes available.")
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(width)

	content := m.renderPaneContent()

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, tabRow, contentBox.Render(content)))
}

func (m Model) renderPaneContent() string {
	if m.selectedPaneID == "" {
		return "Select a pane to view its output."
	}
	content, ok := m.paneContents[m.selectedPaneID]
	if !ok {
		return "Loading pane content..."
	}
	if content.err != nil {
		return fmt.Sprintf("Error capturing pane: %v", content.err)
	}
	text := strings.TrimRight(content.text, "\n")
	if strings.TrimSpace(text) == "" {
		return "[pane is empty]"
	}
	if !content.fetched.IsZero() {
		return fmt.Sprintf("%s\n\n-- captured %s ago --", text, humanizeDuration(time.Since(content.fetched)))
	}
	return text
}

func (m Model) renderStatus() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	last := "never"
	if !m.lastUpdated.IsZero() {
		last = fmt.Sprintf("%s ago", humanizeDuration(time.Since(m.lastUpdated)))
	}
	errPart := ""
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
		errPart = errStyle.Render("Error: " + m.err.Error())
	}
	hiddenCount := len(m.hidden)
	hiddenPart := ""
	if hiddenCount > 0 {
		hiddenPart = fmt.Sprintf(" | hidden: %d", hiddenCount)
	}
	controls := " | keys: j/k←/→ move | [ ] reorder | h hide | H show | tab focus"
	info := fmt.Sprintf("focus: %s | last refresh: %s%s%s", m.focusString(), last, hiddenPart, controls)
	if errPart != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, style.Render(info), lipgloss.NewStyle().PaddingLeft(2).Render(errPart))
	}
	return style.Render(info)
}

func (m Model) focusString() string {
	switch m.focus {
	case focusSidebar:
		return "sessions"
	case focusPanes:
		return "panes"
	default:
		return "unknown"
	}
}

func humanizeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format(time.RFC3339)
}

func humanizeDuration(d time.Duration) string {
	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

func scheduleTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
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

func fetchPaneContentCmd(client *tmux.Client, paneID string, lines int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		text, err := client.CapturePane(ctx, paneID, lines)
		if err != nil {
			return paneContentMsg{paneID: paneID, err: err}
		}
		return paneContentMsg{paneID: paneID, text: text}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
