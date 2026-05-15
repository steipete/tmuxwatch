package zone

import tea "charm.land/bubbletea/v2"

var DefaultManager *Manager

func NewGlobal() {
	if DefaultManager != nil {
		return
	}
	DefaultManager = New()
}

func Close() {
	DefaultManager.checkInitialized()
	DefaultManager.Close()
}

func SetEnabled(v bool) {
	DefaultManager.checkInitialized()
	DefaultManager.SetEnabled(v)
}

func Enabled() bool {
	DefaultManager.checkInitialized()
	return DefaultManager.Enabled()
}

func NewPrefix() string {
	DefaultManager.checkInitialized()
	return DefaultManager.NewPrefix()
}

func Mark(id, v string) string {
	DefaultManager.checkInitialized()
	return DefaultManager.Mark(id, v)
}

func Clear(id string) {
	DefaultManager.checkInitialized()
	DefaultManager.Clear(id)
}

func Get(id string) *ZoneInfo {
	DefaultManager.checkInitialized()
	return DefaultManager.Get(id)
}

func Scan(v string) string {
	DefaultManager.checkInitialized()
	return DefaultManager.Scan(v)
}

func AnyInBounds(model tea.Model, mouse tea.MouseMsg) {
	DefaultManager.checkInitialized()
	DefaultManager.AnyInBounds(model, mouse)
}

func AnyInBoundsAndUpdate(model tea.Model, mouse tea.MouseMsg) (tea.Model, tea.Cmd) {
	DefaultManager.checkInitialized()
	return DefaultManager.AnyInBoundsAndUpdate(model, mouse)
}
