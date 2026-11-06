// File state_test.go validates tab and collapse state helpers.
package ui

import (
	"strings"
	"testing"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// TestTabTitles ensures detail tabs appear alongside the overview tab.
func TestTabTitles(t *testing.T) {
	t.Parallel()

	m := &Model{}
	titles := m.tabTitles()
	if len(titles) != 1 || titles[0] != "Overview" {
		t.Fatalf("tabTitles() = %v, want [Overview]", titles)
	}

	m.sessions = []tmux.Session{{ID: "$1", Name: "dev"}}
	titles = m.tabTitles()
	if len(titles) != 2 {
		t.Fatalf("tabTitles() length = %d, want 2", len(titles))
	}
	if titles[1] != "dev" {
		t.Fatalf("session tab title = %q, want dev", titles[1])
	}
}

// TestRenderTabBarWrapsWhenNarrow ensures the tab bar flows into multiple lines when constrained.
func TestRenderTabBarWrapsWhenNarrow(t *testing.T) {
	t.Parallel()

	m := &Model{
		sessions: []tmux.Session{
			{ID: "s1", Name: "alpha"},
			{ID: "s2", Name: "beta"},
			{ID: "s3", Name: "gamma"},
			{ID: "s4", Name: "delta"},
		},
		zonePrefix: "component",
	}
	rendered := m.renderTabBar(12)
	if strings.Count(rendered, "\n") < 1 {
		t.Fatalf("expected multiline tab bar, got %q", rendered)
	}
}

// TestSetActiveTabUpdatesViewMode confirms view mode tracks the active tab.
func TestSetActiveTabUpdatesViewMode(t *testing.T) {
	t.Parallel()

	m := &Model{
		sessions: []tmux.Session{{ID: "s1", Name: "dev"}},
	}
	m.tabTitles()
	m.setActiveTab(0)
	if m.viewMode != viewModeOverview {
		t.Fatalf("viewMode = %v, want overview", m.viewMode)
	}

	m.setActiveTab(1)
	if m.viewMode != viewModeDetail {
		t.Fatalf("setActiveTab(1) should switch to detail, got %v", m.viewMode)
	}

	m.setActiveTab(0)
	if m.viewMode != viewModeOverview {
		t.Fatalf("setActiveTab(0) should return to overview, got %v", m.viewMode)
	}
}

func TestEnterDetailSkipsAHiddenSession(t *testing.T) {
	t.Parallel()

	m := &Model{
		sessions: []tmux.Session{
			{ID: "s1", Name: "one"},
			{ID: "s2", Name: "two"},
		},
		hidden: map[string]struct{}{"s2": {}},
	}
	m.focusedSession = "s1"
	m.enterDetail("s2")
	if m.focusedSession != "s1" || m.detailSession != "" || m.viewMode != viewModeOverview || m.activeTab != 0 {
		t.Fatalf("hidden detail focus=%q detail=%q view=%v tab=%d", m.focusedSession, m.detailSession, m.viewMode, m.activeTab)
	}

	visible := &Model{
		sessions: []tmux.Session{
			{ID: "s1", Name: "one"},
			{ID: "s2", Name: "two"},
		},
	}
	visible.enterDetail("s2")
	if visible.focusedSession != "s2" || visible.detailSession != "s2" || visible.activeTab != 2 || visible.viewMode != viewModeDetail {
		t.Fatalf("visible detail focus=%q detail=%q view=%v tab=%d", visible.focusedSession, visible.detailSession, visible.viewMode, visible.activeTab)
	}
}

// TestToggleCollapsed ensures session collapse state toggles predictably.
func TestToggleCollapsed(t *testing.T) {
	t.Parallel()

	m := &Model{collapsed: make(map[string]struct{})}
	if m.isCollapsed("s1") {
		t.Fatal("expected s1 to be expanded")
	}
	m.toggleCollapsed("s1")
	if !m.isCollapsed("s1") {
		t.Fatal("expected s1 to be collapsed")
	}
	m.toggleCollapsed("s1")
	if m.isCollapsed("s1") {
		t.Fatal("expected s1 to be expanded after toggle")
	}
}
