package zone

import (
	"sort"

	tea "charm.land/bubbletea/v2"
)

type MsgZoneInBounds struct {
	Zone  *ZoneInfo
	Event tea.MouseMsg
}

func (m *Manager) findInBounds(mouse tea.MouseMsg) []*ZoneInfo {
	m.zoneMu.RLock()
	defer m.zoneMu.RUnlock()

	keys := make([]string, 0, len(m.zones))
	for id := range m.zones {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	zones := make([]*ZoneInfo, 0, len(keys))
	for _, id := range keys {
		zone := m.zones[id]
		if zone.InBounds(mouse) {
			zones = append(zones, zone)
		}
	}
	return zones
}

func (m *Manager) IDsInBounds(mouse tea.MouseMsg) []string {
	m.zoneMu.RLock()
	defer m.zoneMu.RUnlock()

	ids := make([]string, 0)
	for id, zone := range m.zones {
		if zone.InBounds(mouse) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (m *Manager) AnyInBoundsAndUpdate(model tea.Model, mouse tea.MouseMsg) (tea.Model, tea.Cmd) {
	zones := m.findInBounds(mouse)
	cmds := make([]tea.Cmd, len(zones))
	for i, zone := range zones {
		model, cmds[i] = model.Update(MsgZoneInBounds{Zone: zone, Event: mouse})
	}
	return model, tea.Batch(cmds...)
}

func (m *Manager) AnyInBounds(model tea.Model, mouse tea.MouseMsg) {
	for _, zone := range m.findInBounds(mouse) {
		_, _ = model.Update(MsgZoneInBounds{Zone: zone, Event: mouse})
	}
}
