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
	minPreviewHeight    = 6
	cardPadding         = 1
	closeLabel          = "[x]"
	scrollStep          = 3
	pulseDuration       = 1500 * time.Millisecond
	quitChordWindow     = 600 * time.Millisecond
	borderColorBase     = "62"
	borderColorFocus    = "212"
	borderColorPulse    = "213"
	borderColorExitFail = "203"
	borderColorExitOK   = "36"
	headerColorBase     = "249"
	headerColorFocus    = "212"
	headerColorPulse    = "219"
	headerColorExitFail = "203"
	headerColorExitOK   = "37"
)

type (
	snapshotMsg    struct{ snapshot tmux.Snapshot }
	paneContentMsg struct {
		sessionID string
		paneID    string
		text      string
		err       error
	}
	errMsg        struct{ err error }
	tickMsg       struct{}
	searchBlurMsg struct{}
)

type sessionPreview struct {
	viewport    *viewport.Model
	paneID      string
	lastContent string
	lastChanged time.Time
}

type cardBounds struct {
	sessionID  string
	top        int
	bottom     int
	left       int
	right      int
	closeLeft  int
	closeRight int
}

// Model owns the Bubble Tea state machine.
type Model struct {
	client       *tmux.Client
	pollInterval time.Duration

	width  int
	height int

	sessions []tmux.Session

	previews map[string]*sessionPreview
	hidden   map[string]struct{}

	searchInput textinput.Model
	searching   bool
	searchQuery string

	focusedSession string
	cardLayout     []cardBounds

	lastUpdated time.Time
	err         error
	inflight    bool

	cachedStatus string
	lastCtrlC    time.Time
}

// NewModel builds a Model with defaults.
func NewModel(client *tmux.Client, poll time.Duration) *Model {
	if poll <= 0 {
		poll = defaultPollInterval
	}
	ti := textinput.New()
	ti.Placeholder = "filter sessions, windows, panes"
	ti.CharLimit = 256
	ti.Prompt = "/ "
	return &Model{
		client:       client,
		pollInterval: poll,
		previews:     make(map[string]*sessionPreview),
		hidden:       make(map[string]struct{}),
		searchInput:  ti,
		cardLayout:   make([]cardBounds, 0),
		inflight:     true,
	}
}

// Init starts initial polling.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(fetchSnapshotCmd(m.client), scheduleTick(m.pollInterval))
}

// Update processes Bubble Tea messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePreviewDimensions(m.filteredSessionCount())
	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchKey(msg)
		}
		if handled, cmd := m.handleGlobalKey(msg); handled {
			return m, cmd
		}
		if handled, cmd := m.handleFocusedKey(msg); handled {
			return m, cmd
		}
	case tea.MouseMsg:
		return m.handleMouse(msg)
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
		if preview, ok := m.previews[msg.sessionID]; ok && preview.paneID == msg.paneID {
			var content string
			if msg.err != nil {
				content = "Pane capture error: " + msg.err.Error()
			} else {
				content = strings.TrimRight(msg.text, "\n")
			}
			if content != preview.lastContent {
				preview.viewport.SetContent(content)
				preview.lastContent = content
				preview.lastChanged = time.Now()
				preview.viewport.GotoBottom()
			}
		}
	case tickMsg:
		if m.inflight {
			return m, nil
		}
		m.inflight = true
		return m, fetchSnapshotCmd(m.client)
	}
	return m, nil
}

func (m *Model) handleGlobalKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "/", "ctrl+f":
		m.resetCtrlC()
		m.searching = true
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.CursorEnd()
		return true, nil
	case "esc":
		if m.searchQuery != "" {
			m.resetCtrlC()
			m.searchQuery = ""
			m.updatePreviewDimensions(m.filteredSessionCount())
			return true, nil
		}
		return false, nil
	case "H":
		if len(m.hidden) > 0 {
			m.resetCtrlC()
			m.hidden = make(map[string]struct{})
			m.updatePreviewDimensions(m.filteredSessionCount())
		}
		return true, nil
	case "q":
		m.resetCtrlC()
		return true, tea.Quit
	}
	return false, nil
}

func (m *Model) handleFocusedKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.focusedSession == "" {
		return false, nil
	}
	preview, ok := m.previews[m.focusedSession]
	if !ok {
		return false, nil
	}
	pane, paneOK := m.paneFor(m.focusedSession)
	switch msg.String() {
	case "up":
		m.resetCtrlC()
		preview.viewport.ScrollUp(1)
		return true, nil
	case "down":
		m.resetCtrlC()
		preview.viewport.ScrollDown(1)
		return true, nil
	case "pgup":
		m.resetCtrlC()
		preview.viewport.PageUp()
		return true, nil
	case "pgdown":
		m.resetCtrlC()
		preview.viewport.PageDown()
		return true, nil
	case "ctrl+u":
		m.resetCtrlC()
		preview.viewport.ScrollUp(scrollStep)
		return true, nil
	case "ctrl+d":
		m.resetCtrlC()
		preview.viewport.ScrollDown(scrollStep)
		return true, nil
	case "g":
		m.resetCtrlC()
		preview.viewport.GotoTop()
		return true, nil
	case "G":
		m.resetCtrlC()
		preview.viewport.GotoBottom()
		return true, nil
	case "ctrl+c":
		now := time.Now()
		if !paneOK || pane.Dead || preview.paneID == "" {
			return true, tea.Quit
		}
		cmd := sendKeysCmd(m.client, preview.paneID, "C-c")
		if !m.lastCtrlC.IsZero() && now.Sub(m.lastCtrlC) < quitChordWindow {
			m.resetCtrlC()
			return true, tea.Batch(cmd, tea.Quit)
		}
		m.lastCtrlC = now
		return true, cmd
	}

	keys, ok := tmuxKeysFrom(msg)
	if !ok || preview.paneID == "" {
		m.resetCtrlC()
		return false, nil
	}
	m.resetCtrlC()
	return true, sendKeysCmd(m.client, preview.paneID, keys...)
}

func tmuxKeysFrom(msg tea.KeyMsg) ([]string, bool) {
	switch msg.Type {
	case tea.KeyEnter:
		return []string{"Enter"}, true
	case tea.KeyTab:
		return []string{"Tab"}, true
	case tea.KeySpace:
		return []string{" "}, true
	case tea.KeyBackspace:
		return []string{"BSpace"}, true
	case tea.KeyDelete:
		return []string{"Delete"}, true
	case tea.KeyEsc:
		return []string{"Escape"}, true
	case tea.KeyRunes:
		if msg.Alt {
			return nil, false
		}
		return []string{string(msg.Runes)}, true
	}
	return nil, false
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if len(m.cardLayout) == 0 {
		return m, nil
	}
	card, ok := m.cardAt(msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	preview := m.previews[card.sessionID]
	switch {
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		if preview != nil {
			_ = preview.viewport.ScrollDown(scrollStep)
		}
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		if preview != nil {
			_ = preview.viewport.ScrollUp(scrollStep)
		}
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		if msg.Y >= card.top && msg.Y <= card.top+1 && msg.X >= card.closeLeft && msg.X <= card.closeRight {
			m.hidden[card.sessionID] = struct{}{}
			if m.focusedSession == card.sessionID {
				m.focusedSession = ""
			}
			delete(m.previews, card.sessionID)
			m.resetCtrlC()
			m.updatePreviewDimensions(m.filteredSessionCount())
			return m, nil
		}
		m.focusedSession = card.sessionID
		m.resetCtrlC()
		if preview != nil {
			preview.viewport.GotoBottom()
		}
	}
	return m, nil
}

// View renders the UI.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	var sections []string
	title := renderTitleBar()
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

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.isHidden(session.ID) {
			continue
		}
		active[session.ID] = struct{}{}
		window, ok := activeWindow(session)
		if !ok {
			continue
		}
		pane, ok := activePane(window)
		if !ok {
			continue
		}
		preview := m.previews[session.ID]
		if preview == nil {
			vp := viewport.New(0, minPreviewHeight)
			vp.MouseWheelEnabled = false
			vp.MouseWheelDelta = scrollStep
			preview = &sessionPreview{viewport: &vp, lastChanged: time.Now()}
			m.previews[session.ID] = preview
		}
		if preview.paneID != pane.ID {
			preview.viewport.SetContent("")
			preview.paneID = pane.ID
			preview.lastContent = ""
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

func (m *Model) filteredSessions() []tmux.Session {
	var out []tmux.Session
	query := strings.ToLower(m.searchQuery)
	for _, session := range m.sessions {
		if m.isHidden(session.ID) {
			continue
		}
		if query == "" || sessionMatches(session, query) {
			out = append(out, session)
		}
	}
	return out
}

func (m *Model) filteredSessionCount() int {
	return len(m.filteredSessions())
}

func (m *Model) renderSessionPreviews(offset int) string {
	sessions := m.filteredSessions()
	m.cardLayout = m.cardLayout[:0]
	if len(sessions) == 0 {
		return ""
	}

	baseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColorBase)).
		Padding(0, cardPadding)

	cardLeft := 0
	var rendered []string
	currentY := offset + 1
	now := time.Now()

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

		innerWidth := max(20, preview.viewport.Width)
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

		rendered = append(rendered, card)

		height := lipgloss.Height(card)
		width := lipgloss.Width(card)
		closeWidth := len(closeLabel)
		cardTop := currentY
		currentY += height
		left := cardLeft
		right := left + width - 1
		closeRight := right - 1 - cardPadding
		if closeRight > right {
			closeRight = right
		}
		if closeRight < left {
			closeRight = left
		}
		closeLeft := max(left, closeRight-closeWidth+1)
		bounds := cardBounds{
			sessionID:  session.ID,
			top:        cardTop,
			bottom:     currentY - 1,
			left:       left,
			right:      right,
			closeLeft:  closeLeft,
			closeRight: closeRight,
		}
		m.cardLayout = append(m.cardLayout, bounds)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

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

func (m *Model) updatePreviewDimensions(count int) {
	if count <= 0 || m.width <= 0 || m.height <= 0 {
		return
	}
	frameHeight := 3
	internalHeight := (m.height / count) - frameHeight
	if internalHeight < minPreviewHeight {
		internalHeight = minPreviewHeight
	}
	innerWidth := max(20, m.width-(cardPadding*2+2))
	for _, preview := range m.previews {
		preview.viewport.Width = innerWidth
		preview.viewport.Height = internalHeight
	}
}

func (m *Model) renderStatus() string {
	content := m.buildStatusLine()
	if content == m.cachedStatus {
		return m.cachedStatus
	}
	m.cachedStatus = content
	return content
}

func (m *Model) buildStatusLine() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 2)
	last := "never"
	if !m.lastUpdated.IsZero() {
		last = fmt.Sprintf("%s ago", coarseDuration(time.Since(m.lastUpdated)))
	}
	errPart := ""
	if m.err != nil {
		errPart = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("Error: " + m.err.Error())
	}
	filterPart := ""
	if m.searchQuery != "" {
		filterPart = fmt.Sprintf(" | filter: %s", m.searchQuery)
	}
	focusPart := ""
	if m.focusedSession != "" {
		focusPart = fmt.Sprintf(" | focused: %s", m.focusedSession)
	}
	info := fmt.Sprintf("sessions: %d | last refresh: %s%s%s", len(m.sessions), last, filterPart, focusPart)
	help := "mouse: click focus, scroll logs, close [x] · keys: / search, H show hidden, q quit"
	status := lipgloss.JoinHorizontal(lipgloss.Left, style.Render(info), lipgloss.NewStyle().PaddingLeft(2).Render(help))
	if errPart != "" {
		status = lipgloss.JoinHorizontal(lipgloss.Left, status, lipgloss.NewStyle().PaddingLeft(2).Render(errPart))
	}
	return status
}

func coarseDuration(d time.Duration) string {
	switch {
	case d < 5*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()/5)*5)
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
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

func renderTitleBar() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 2)
	return style.Render("tmuxwatch ▸ live tmux monitor")
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

func (m *Model) isHidden(id string) bool {
	_, ok := m.hidden[id]
	return ok
}

func (m *Model) cardAt(x, y int) (cardBounds, bool) {
	for _, card := range m.cardLayout {
		if y >= card.top && y <= card.bottom && x >= card.left && x <= card.right {
			return card, true
		}
	}
	return cardBounds{}, false
}

func (m *Model) resetCtrlC() {
	m.lastCtrlC = time.Time{}
}

func (m *Model) paneFor(sessionID string) (tmux.Pane, bool) {
	for _, session := range m.sessions {
		if session.ID == sessionID {
			window, ok := activeWindow(session)
			if !ok {
				return tmux.Pane{}, false
			}
			return activePane(window)
		}
	}
	return tmux.Pane{}, false
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

func sendKeysCmd(client *tmux.Client, paneID string, keys ...string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.SendKeys(ctx, paneID, keys...); err != nil {
			return errMsg{err: err}
		}
		return nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
