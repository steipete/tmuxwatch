// File stale.go encapsulates logic for detecting inactive or dead sessions.
package ui

import "time"

func (m *Model) updateStaleSessions() {
	for k := range m.stale {
		delete(m.stale, k)
	}
	now := time.Now()
	for _, session := range m.sessions {
		if sessionAllPanesDead(session) {
			m.stale[session.ID] = struct{}{}
			continue
		}
		last := sessionLatestActivity(session)
		if last.IsZero() {
			last = session.CreatedAt
		}
		if now.Sub(last) >= staleThreshold {
			m.stale[session.ID] = struct{}{}
		}
	}
}

func (m *Model) isStale(sessionID string) bool {
	_, ok := m.stale[sessionID]
	return ok
}

func (m *Model) staleSessionNames() []string {
	if len(m.stale) == 0 {
		return nil
	}
	names := make([]string, 0, len(m.stale))
	for _, session := range m.sessions {
		if m.isStale(session.ID) {
			names = append(names, session.Name)
		}
	}
	return names
}
