// File model.go defines the core Bubble Tea model structure and shared
// constants that govern tmuxwatch behaviour.
package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

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
	paneVarsMsg struct {
		sessionID string
		paneID    string
		vars      map[string]string
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
	vars        map[string]string
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

// Model owns the Bubble Tea state machine and cached tmux snapshot data.
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

// NewModel builds a Model with defaults and the provided tmux client.
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

// Init starts the initial tmux snapshot fetch and ticking loop.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(fetchSnapshotCmd(m.client), scheduleTick(m.pollInterval))
}
