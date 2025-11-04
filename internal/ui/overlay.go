package ui

import (
	"image"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/cellbuf"
)

func overlayView(base, overlay string, width, height, offsetX, offsetY int) string {
	baseWidth := lipgloss.Width(base)
	baseHeight := countLines(base)

	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := countLines(overlay)

	if width <= 0 {
		width = baseWidth
	}
	width = max(width, offsetX+overlayWidth)
	if width <= 0 {
		width = 1
	}

	if height <= 0 {
		height = baseHeight
	}
	height = max(height, offsetY+overlayHeight)
	if height <= 0 {
		height = 1
	}

	baseBuf := cellbuf.NewBuffer(width, height)
	cellbuf.SetContentRect(baseBuf, base, image.Rect(0, 0, width, height))

	if overlay != "" {
		oBuf := cellbuf.NewBuffer(overlayWidth, overlayHeight)
		cellbuf.SetContentRect(oBuf, overlay, image.Rect(0, 0, overlayWidth, overlayHeight))

		for y := 0; y < overlayHeight; y++ {
			for x := 0; x < overlayWidth; x++ {
				cell := oBuf.Cell(x, y)
				if cell == nil || cell.Empty() {
					continue
				}
				destX := offsetX + x
				destY := offsetY + y
				if destX < 0 || destX >= width || destY < 0 || destY >= height {
					continue
				}
				baseBuf.SetCell(destX, destY, cell.Clone())
			}
		}
	}

	out := cellbuf.Render(baseBuf)
	return strings.ReplaceAll(out, "\r\n", "\n")
}

func countLines(str string) int {
	if str == "" {
		return 1
	}
	return strings.Count(str, "\n") + 1
}
