package ui

import (
	"image"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/cellbuf"
)

func overlayView(base, overlay string, width, height int) string {
	if width <= 0 {
		width = max(lipgloss.Width(base), lipgloss.Width(overlay))
	}
	if width <= 0 {
		width = lipgloss.Width(base)
	}
	if width <= 0 {
		width = 1
	}

	if height <= 0 {
		height = max(countLines(base), countLines(overlay))
	}
	if height <= 0 {
		height = 1
	}

	rect := image.Rect(0, 0, width, height)

	baseBuf := cellbuf.NewBuffer(width, height)
	cellbuf.SetContentRect(baseBuf, base, rect)

	overlayBuf := cellbuf.NewBuffer(width, height)
	cellbuf.SetContentRect(overlayBuf, overlay, rect)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := overlayBuf.Cell(x, y)
			if cell == nil || cell.Empty() || cell.Clear() {
				continue
			}
			// Clone before writing so we don't mutate the overlay buffer.
			baseBuf.SetCell(x, y, cell.Clone())
		}
	}

	out := cellbuf.Render(baseBuf)
	out = strings.ReplaceAll(out, "\r\n", "\n")
	return out
}

func countLines(str string) int {
	if str == "" {
		return 1
	}
	return strings.Count(str, "\n") + 1
}
