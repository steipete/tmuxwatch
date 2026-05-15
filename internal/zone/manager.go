package zone

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	identStart   = '\x1B'
	identBracket = '['
	identEnd     = 'z'
)

var (
	markerCounter int64 = 1000
	prefixCounter int64
)

type Manager struct {
	enabled atomic.Bool

	zoneMu sync.RWMutex
	zones  map[string]*ZoneInfo

	idMu sync.RWMutex
	ids  map[string]string
	rids map[string]string
}

func New() *Manager {
	m := &Manager{
		zones: make(map[string]*ZoneInfo),
		ids:   make(map[string]string),
		rids:  make(map[string]string),
	}
	m.enabled.Store(true)
	return m
}

func (m *Manager) checkInitialized() {
	if m == nil {
		panic("zone manager not initialized")
	}
}

func (m *Manager) Close() {}

func (m *Manager) SetEnabled(enabled bool) {
	m.enabled.Store(enabled)
	if !enabled {
		m.zoneMu.Lock()
		clear(m.zones)
		m.zoneMu.Unlock()
	}
}

func (m *Manager) Enabled() bool {
	return m.enabled.Load()
}

func (m *Manager) NewPrefix() string {
	return "zone_" + strconv.FormatInt(atomic.AddInt64(&prefixCounter, 1), 10) + "__"
}

func (m *Manager) Mark(id, v string) string {
	if !m.Enabled() || id == "" || v == "" {
		return v
	}

	m.idMu.RLock()
	gid := m.ids[id]
	m.idMu.RUnlock()
	if gid != "" {
		return gid + v + gid
	}

	m.idMu.Lock()
	defer m.idMu.Unlock()
	if gid = m.ids[id]; gid != "" {
		return gid + v + gid
	}

	gid = string(identStart) + string(identBracket) + strconv.FormatInt(atomic.AddInt64(&markerCounter, 1), 10) + string(identEnd)
	m.ids[id] = gid
	m.rids[gid] = id
	return gid + v + gid
}

func (m *Manager) Clear(id string) {
	m.zoneMu.Lock()
	delete(m.zones, id)
	m.zoneMu.Unlock()
}

func (m *Manager) Get(id string) *ZoneInfo {
	m.zoneMu.RLock()
	zone := m.zones[id]
	m.zoneMu.RUnlock()
	return zone
}

func (m *Manager) getReverse(id string) string {
	m.idMu.RLock()
	resolved := m.rids[id]
	m.idMu.RUnlock()
	return resolved
}

func (m *Manager) Scan(v string) string {
	scanner := newScanner(m, v, time.Now().Nanosecond())
	scanner.run()

	m.zoneMu.Lock()
	clear(m.zones)
	for rid, info := range scanner.found {
		if id := m.getReverse(rid); id != "" {
			m.zones[id] = info
		}
	}
	m.zoneMu.Unlock()

	return scanner.input
}
