// File model.go defines the core Bubble Tea model structure and shared
// constants that govern tmuxwatch behaviour.
package ui

import (
	"os"
	"time"

	zone "github.com/alexanderbh/bubblezone/v2"
	"github.com/charmbracelet/bubbles/v2/textinput"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

const (
	defaultPollInterval = time.Second
	minPreviewHeight    = 6
	cardPadding         = 1
	closeLabel          = "[x]"
	maximizeLabel       = "[^]"
	restoreLabel        = "[v]"
	collapseLabel       = "[-]"
	expandLabel         = "[+]"
	scrollStep          = 3
	pulseDuration       = 1500 * time.Millisecond
	quitChordWindow     = 600 * time.Millisecond
	staleThreshold      = time.Hour
	minCaptureLines     = 120
	maxCaptureLines     = 800
	captureSlackLines   = 80
	borderColorBase     = "62"
	borderColorFocus    = "212"
	borderColorPulse    = "213"
	borderColorCursor   = "111"
	borderColorHover    = "143"
	borderColorExitFail = "203"
	borderColorExitOK   = "36"
	borderColorStale    = "95"
	headerColorBase     = "249"
	headerColorFocus    = "212"
	headerColorPulse    = "219"
	headerColorCursor   = "111"
	headerColorExitFail = "203"
	headerColorExitOK   = "37"
	headerColorStale    = "103"
)

type viewMode int

const (
	viewModeOverview viewMode = iota
	viewModeDetail
)

type (
	snapshotMsg    struct{ snapshot tmux.Snapshot }
	statusMsg      string
	paneContentMsg struct {
		sessionID string
		paneID    string
		text      string
		err       error
	}
	paneVarsMsg struct {
		sessionID string
		paneID    string
		vars      map[string]string
		err       error
	}
	killSessionsMsg struct{ ids []string }
	errMsg          struct{ err error }
	tickMsg         struct{}
	searchBlurMsg   struct{}
)

type sessionPreview struct {
	viewport    *viewport.Model
	paneID      string
	lastContent string
	lastChanged time.Time
	vars        map[string]string
}

type cardBounds struct {
	sessionID      string
	zoneID         string
	closeZoneID    string
	maximizeZoneID string
	collapseZoneID string
}

type commandItem struct {
	label   string
	enabled bool
	run     func(*Model) tea.Cmd
}

// Model owns the Bubble Tea state machine and cached tmux snapshot data.
type Model struct {
	client       *tmux.Client
	pollInterval time.Duration
	zonePrefix   string

	width  int
	height int

	sessions []tmux.Session

	previews  map[string]*sessionPreview
	hidden    map[string]struct{}
	stale     map[string]struct{}
	collapsed map[string]struct{}

	paletteOpen     bool
	paletteIndex    int
	paletteCommands []commandItem

	searchInput textinput.Model
	searching   bool
	searchQuery string
	toast       *toastState

	focusedSession string
	cardLayout     []cardBounds
	cursorSession  string
	hoveredSession string
	hoveredControl string

	viewMode      viewMode
	detailSession string
	activeTab     int
	cardCols      int
	previewOffset int
	tabSessionIDs []string

	debugMsgs  []tea.Msg
	traceMouse bool
	hostname   string

	lastUpdated time.Time
	err         error
	inflight    bool

	cachedStatus string
	lastCtrlC    time.Time
	lastEsc      time.Time
}

// sessionLabel strips leading sigils from tmux session identifiers for
// friendlier display.
func sessionLabel(id string) string {
	if len(id) > 1 && id[0] == '$' {
		return id[1:]
	}
	return id
}

// NewModel builds a Model with defaults and the provided tmux client.
func NewModel(client *tmux.Client, poll time.Duration, debugMsgs []tea.Msg, traceMouse bool) *Model {
	if poll <= 0 {
		poll = defaultPollInterval
	}
	ti := textinput.New()
	ti.Placeholder = "filter sessions, windows, panes"
	ti.CharLimit = 256
	ti.Prompt = "/ "
	return &Model{
		client:        client,
		pollInterval:  poll,
		zonePrefix:    zone.NewPrefix(),
		previews:      make(map[string]*sessionPreview),
		hidden:        make(map[string]struct{}),
		stale:         make(map[string]struct{}),
		collapsed:     make(map[string]struct{}),
		searchInput:   ti,
		cardLayout:    make([]cardBounds, 0),
		cardCols:      1,
		inflight:      true,
		previewOffset: topPaddingLines,
		debugMsgs:     append([]tea.Msg(nil), debugMsgs...),
		traceMouse:    traceMouse,
		toast:         &toastState{},
		viewMode:      viewModeOverview,
		tabSessionIDs: make([]string, 0),
		hostname:      lookupHostname(),
	}
}

func lookupHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

// Init starts the initial tmux snapshot fetch and ticking loop.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		fetchSnapshotCmd(m.client),
		scheduleTick(m.pollInterval),
	}
	for _, msg := range m.debugMsgs {
		cmds = append(cmds, emitMsg(msg))
	}
	m.debugMsgs = nil
	return tea.Batch(cmds...)
}
