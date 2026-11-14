// File handlers_test.go validates keyboard and mouse handlers used by the UI.
package ui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	zone "github.com/steipete/tmuxwatch/internal/zone"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestTmuxKeysFrom ensures Bubble Tea key messages map to tmux key strings.
func TestTmuxKeysFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		msg    tea.KeyMsg
		want   []string
		expect bool
	}{
		{name: "enter", msg: tea.KeyPressMsg{Code: tea.KeyEnter}, want: []string{"Enter"}, expect: true},
		{name: "space", msg: tea.KeyPressMsg{Code: tea.KeySpace}, want: []string{" "}, expect: true},
		{name: "alt runes rejected", msg: tea.KeyPressMsg{Text: "a", Code: 'a', Mod: tea.ModAlt}, expect: false},
		{name: "runes ok", msg: tea.KeyPressMsg{Text: "a", Code: 'a'}, want: []string{"a"}, expect: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := tmuxKeysFrom(tt.msg)
			if ok != tt.expect {
				t.Fatalf("tmuxKeysFrom ok = %v, want %v", ok, tt.expect)
			}
			if !tt.expect {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("tmuxKeysFrom len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("tmuxKeysFrom got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestTabIndexFromZoneIDs checks zone identifiers convert to tab indexes.
func TestTabIndexFromZoneIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ids      []string
		wantIdx  int
		expectOK bool
	}{
		{name: "no match", ids: []string{"foo"}},
		{name: "root id", ids: []string{"component###tab:0"}, wantIdx: 0, expectOK: true},
		{name: "multiple ids", ids: []string{"component###tab:1", "other"}, wantIdx: 1, expectOK: true},
		{name: "parse failure", ids: []string{"component###tab:notint"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idx, ok := tabIndexFromZoneIDs(tt.ids)
			if ok != tt.expectOK {
				t.Fatalf("tabIndexFromZoneIDs ok = %v, want %v", ok, tt.expectOK)
			}
			if !ok {
				return
			}
			if idx != tt.wantIdx {
				t.Fatalf("tabIndexFromZoneIDs = %d, want %d", idx, tt.wantIdx)
			}
		})
	}
}

// TestHandleTabMouseClick confirms tab clicks update the active tab state.
func TestHandleTabMouseClick(t *testing.T) {
	zone.NewGlobal()
	m := &Model{
		sessions:   []tmux.Session{{ID: "sess", Name: "sess"}},
		hidden:     make(map[string]struct{}),
		stale:      make(map[string]struct{}),
		collapsed:  make(map[string]struct{}),
		zonePrefix: zone.NewPrefix(),
		width:      80,
	}
	bar := m.renderTabBar(80)
	_ = zone.Scan(bar)
	tabID := m.zonePrefix + "###tab:1"
	var info *zone.ZoneInfo
	deadline := time.Now().Add(50 * time.Millisecond)
	for info == nil && time.Now().Before(deadline) {
		info = zone.Get(tabID)
		time.Sleep(time.Millisecond)
	}
	if info == nil {
		t.Fatal("expected session tab zone to be registered")
	}

	msg := tea.MouseClickMsg{X: info.StartX, Y: info.StartY, Button: tea.MouseLeft}
	ids := zone.DefaultManager.IDsInBounds(msg)
	if len(ids) == 0 {
		ids = zone.DefaultManager.IDsInBounds(tea.MouseMotionMsg{X: info.StartX, Y: info.StartY})
	}
	if len(ids) == 0 {
		t.Fatalf("expected zone ids at (%d,%d)", info.StartX, info.StartY)
	}
	handled, cmd := m.handleTabMouse(msg)
	if !handled {
		t.Fatal("expected tab click to be handled")
	}
	if cmd != nil {
		t.Fatal("expected no command from tab click")
	}
	if m.activeTab != 1 {
		t.Fatalf("activeTab = %d, want 1", m.activeTab)
	}
}

// TestHandleMouseHoverSetsState ensures motion tracking highlights the card.
func TestHandleMouseHoverSetsState(t *testing.T) {
	zone.DefaultManager = zone.New()
	m := &Model{
		previews: map[string]*sessionPreview{
			"s1": {viewport: func() *viewport.Model {
				vp := viewportFor(innerDimension{width: 60, height: 20})
				return &vp
			}()},
		},
		sessions: []tmux.Session{{
			ID: "s1",
			Windows: []tmux.Window{{
				Active: true,
				Panes:  []tmux.Pane{{ID: "%1", Active: true}},
			}},
		}},
		hidden:     make(map[string]struct{}),
		stale:      make(map[string]struct{}),
		collapsed:  make(map[string]struct{}),
		zonePrefix: zone.NewPrefix(),
		width:      120,
		height:     50,
	}

	view := m.renderSessionPreviews(0)
	_ = zone.Scan(view)

	card := m.cardLayout[0]
	var info *zone.ZoneInfo
	deadline := time.Now().Add(500 * time.Millisecond)
	for info == nil && time.Now().Before(deadline) {
		info = zone.Get(card.zoneID)
		time.Sleep(2 * time.Millisecond)
	}
	if info == nil {
		t.Fatal("expected zone info to be registered")
	}

	m.handleMouse(tea.MouseMotionMsg{X: info.StartX, Y: info.StartY})
	if m.hoveredSession != "s1" {
		t.Fatalf("hoveredSession = %q, want s1", m.hoveredSession)
	}
}

// TestFocusedEnterSendsKeys ensures Enter on an already focused session is
// forwarded with tmux send-keys instead of being swallowed as a focus toggle.
func TestFocusedEnterSendsKeys(t *testing.T) {
	t.Parallel()

	client, logPath := newRecordingTmux(t)
	const sessionID = "$1"
	const paneID = "%9"
	m := focusedEnterModel(client, sessionID, paneID)
	m.focusedSession = sessionID
	m.cursorSession = sessionID

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("focused Enter returned no command; send-keys was not called")
	}
	if msg := cmd(); msg != nil {
		t.Fatalf("focused Enter command result = %#v, want nil", msg)
	}
	got := readRecording(t, logPath)
	want := "send-keys -t " + paneID + " Enter\n"
	if got != want {
		t.Fatalf("tmux invocation = %q, want %q", got, want)
	}
}

// TestUnfocusedEnterFocusesWithoutSendKeys ensures the first Enter focuses the
// cursor session and does not type into the pane.
func TestUnfocusedEnterFocusesWithoutSendKeys(t *testing.T) {
	t.Parallel()

	client, logPath := newRecordingTmux(t)
	const sessionID = "$1"
	const paneID = "%9"
	m := focusedEnterModel(client, sessionID, paneID)
	m.cursorSession = sessionID

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.focusedSession != sessionID {
		t.Fatalf("focusedSession = %q, want %q", m.focusedSession, sessionID)
	}
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(errMsg); ok {
				t.Fatalf("unfocused Enter command failed: %#v", msg)
			}
		}
	}
	got := readRecording(t, logPath)
	if strings.Contains(got, "send-keys") {
		t.Fatalf("unfocused Enter invoked send-keys: %q", got)
	}
}

func focusedEnterModel(client *tmux.Client, sessionID, paneID string) *Model {
	vp := viewportFor(innerDimension{width: 40, height: 8})
	return &Model{
		client: client,
		sessions: []tmux.Session{{
			ID: sessionID,
			Windows: []tmux.Window{{
				Active: true,
				Panes:  []tmux.Pane{{ID: paneID, Active: true}},
			}},
		}},
		previews: map[string]*sessionPreview{
			sessionID: {viewport: &vp, paneID: paneID},
		},
		hidden:    make(map[string]struct{}),
		stale:     make(map[string]struct{}),
		collapsed: make(map[string]struct{}),
	}
}

func newRecordingTmux(t *testing.T) (*tmux.Client, string) {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls.log")
	binPath := filepath.Join(dir, "tmux")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + strconv.Quote(logPath) + "\n"
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake tmux: %v", err)
	}
	client, err := tmux.NewClient(binPath)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, logPath
}

func readRecording(t *testing.T, logPath string) string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read fake tmux log: %v", err)
	}
	return string(data)
}

// TestControlUnderPointer identifies active control zones for hover styling.

func TestEscLeavesDetailFromALaterTab(t *testing.T) {
	m := &Model{
		sessions: []tmux.Session{
			{ID: "s1", Name: "one"},
			{ID: "s2", Name: "two"},
		},
	}
	m.tabTitles()
	for _, tab := range []int{1, 2} {
		m.setActiveTab(tab)
		if m.viewMode != viewModeDetail || m.activeTab != tab {
			t.Fatalf("tab %d setup view=%v active=%d", tab, m.viewMode, m.activeTab)
		}
		handled, cmd := m.handleGlobalKey(tea.KeyPressMsg{Code: tea.KeyEscape})
		if !handled || cmd != nil {
			t.Fatalf("tab %d handled=%v cmd=%v", tab, handled, cmd)
		}
		if m.viewMode != viewModeOverview || m.activeTab != 0 {
			t.Fatalf("tab %d after esc view=%v active=%d", tab, m.viewMode, m.activeTab)
		}
	}
}
