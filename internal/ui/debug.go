// File debug.go collects optional tracing helpers for mouse and layout data.
package ui

import (
	"fmt"
	"os"

	zone "github.com/alexanderbh/bubblezone/v2"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// logMouseEvent prints mouse hit testing diagnostics when trace mode is on.
func (m *Model) logMouseEvent(msg tea.MouseMsg, card cardBounds, ok bool) {
	if !m.traceMouse {
		return
	}
	result := "miss"
	if ok {
		if info := zone.Get(card.zoneID); info != nil {
			result = fmt.Sprintf("session=%s bounds=[x=%d..%d y=%d..%d]", card.sessionID, info.StartX, info.EndX, info.StartY, info.EndY)
		} else {
			result = fmt.Sprintf("session=%s bounds=<unknown>", card.sessionID)
		}
	}
	mouse := msg.Mouse()
	fmt.Fprintf(os.Stderr, "[mouse] event=%s button=%v pos=(%d,%d) -> %s\n", msg, mouse.Button, mouse.X, mouse.Y, result)
}

// logCardLayout dumps the latest card layout boundaries for bubblezone debug.
func (m *Model) logCardLayout() {
	if !m.traceMouse {
		return
	}
	fmt.Fprintln(os.Stderr, "[debug] card layout:")
	for i, card := range m.cardLayout {
		info := zone.Get(card.zoneID)
		closeInfo := zone.Get(card.closeZoneID)
		if info != nil {
			fmt.Fprintf(os.Stderr, "  [%d] session=%s bounds=[x=%d..%d y=%d..%d]\n",
				i, card.sessionID, info.StartX, info.EndX, info.StartY, info.EndY)
		} else {
			fmt.Fprintf(os.Stderr, "  [%d] session=%s bounds=<unknown>\n", i, card.sessionID)
		}
		if closeInfo != nil {
			fmt.Fprintf(os.Stderr, "     close=[x=%d..%d y=%d..%d]\n", closeInfo.StartX, closeInfo.EndX, closeInfo.StartY, closeInfo.EndY)
		}
	}
}
