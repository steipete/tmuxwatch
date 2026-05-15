package zone

import tea "charm.land/bubbletea/v2"

type ZoneInfo struct {
	Id        string
	iteration int

	StartX int
	StartY int
	EndX   int
	EndY   int
}

func (z *ZoneInfo) IsZero() bool {
	return z == nil || z.Id == ""
}

func (z *ZoneInfo) InBounds(e tea.MouseMsg) bool {
	if z.IsZero() || z.StartX > z.EndX || z.StartY > z.EndY {
		return false
	}

	mouse := e.Mouse()
	return mouse.X >= z.StartX && mouse.Y >= z.StartY && mouse.X <= z.EndX && mouse.Y <= z.EndY
}

func (z *ZoneInfo) Pos(msg tea.MouseMsg) (x, y int) {
	if z.IsZero() || !z.InBounds(msg) {
		return -1, -1
	}

	mouse := msg.Mouse()
	return mouse.X - z.StartX, mouse.Y - z.StartY
}
