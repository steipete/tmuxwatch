package ui

import (
	"testing"
	"time"

	zone "github.com/alexanderbh/bubblezone/v2"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

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

func TestHandleTabMouseClick(t *testing.T) {
	t.Parallel()

	zone.NewGlobal()
	id := "component###tab:1"
	view := zone.Mark(id, "T")
	_ = zone.Scan(view)
	var info *zone.ZoneInfo
	deadline := time.Now().Add(50 * time.Millisecond)
	for info == nil && time.Now().Before(deadline) {
		info = zone.Get(id)
	}
	if info == nil {
		t.Fatal("expected zone info to be registered")
	}

	m := &Model{
		tabRenderer:   newTabRenderer(),
		detailSession: "sess",
		sessions:      []tmux.Session{{ID: "sess", Name: "sess"}},
	}
	m.setActiveTab(0)

	msg := tea.MouseClickMsg{X: info.StartX, Y: info.StartY, Button: tea.MouseLeft}
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
