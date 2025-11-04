package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) logMouseEvent(msg tea.MouseMsg, card cardBounds, ok bool) {
	if !m.traceMouse {
		return
	}
	result := "miss"
	if ok {
		x0 := card.left
		x1 := card.left + card.width - 1
		y0 := card.top
		y1 := card.top + card.height - 1
		result = fmt.Sprintf("session=%s bounds=[x=%d..%d y=%d..%d]", card.sessionID, x0, x1, y0, y1)
	}
	fmt.Fprintf(os.Stderr, "[mouse] action=%v button=%v pos=(%d,%d) -> %s\n", msg.Action, msg.Button, msg.X, msg.Y, result)
}

func (m *Model) logCardLayout() {
	if !m.traceMouse {
		return
	}
	fmt.Fprintln(os.Stderr, "[debug] card layout:")
	for i, card := range m.cardLayout {
		x0 := card.left
		x1 := card.left + card.width - 1
		y0 := card.top
		y1 := card.top + card.height - 1
		fmt.Fprintf(os.Stderr, "  [%d] session=%s bounds=[x=%d..%d y=%d..%d] close=[x=%d..%d y=%d..%d]\n",
			i,
			card.sessionID,
			x0, x1,
			y0, y1,
			card.closeLeft, card.closeRight,
			card.closeTop, card.closeBottom)
	}
}
