package zone

import (
	"unicode/utf8"

	"github.com/muesli/ansi"
)

const eof = 1

type stateFn func(*scanner) stateFn

type scanner struct {
	manager   *Manager
	enabled   bool
	iteration int

	input string
	pos   int
	start int
	width int

	newlines    int
	lastNewline int

	tracked map[string]*ZoneInfo
	found   map[string]*ZoneInfo
}

func newScanner(m *Manager, input string, iteration int) *scanner {
	return &scanner{
		manager:   m,
		enabled:   m.Enabled(),
		iteration: iteration,
		input:     input,
		tracked:   make(map[string]*ZoneInfo),
		found:     make(map[string]*ZoneInfo),
	}
}

func (s *scanner) run() {
	for state := scanMain; state != nil; {
		state = state(s)
	}
}

func (s *scanner) emit() {
	if !s.enabled {
		s.input = s.input[:s.start] + s.input[s.pos:]
		s.pos = s.start
		return
	}

	rid := s.input[s.start:s.pos]
	if item, ok := s.tracked[rid]; ok {
		item.EndX = ansi.PrintableRuneWidth(s.input[s.lastNewline:s.start]) - 1
		item.EndY = s.newlines
		s.found[rid] = item
		delete(s.tracked, rid)
	} else {
		s.tracked[rid] = &ZoneInfo{
			Id:        s.manager.getReverse(rid),
			iteration: s.iteration,
			StartX:    ansi.PrintableRuneWidth(s.input[s.lastNewline:s.start]),
			StartY:    s.newlines,
		}
	}

	s.input = s.input[:s.start] + s.input[s.pos:]
	s.pos = s.start
}

func (s *scanner) next() rune {
	if s.pos >= len(s.input) {
		s.width = 0
		return eof
	}

	r, width := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = width
	s.pos += width
	return r
}

func (s *scanner) backup() {
	s.pos -= s.width
}

func (s *scanner) peek() rune {
	r := s.next()
	s.backup()
	return r
}

func scanMain(s *scanner) stateFn {
	switch r := s.next(); r {
	case eof:
		return nil
	case '\n':
		s.newlines++
		s.lastNewline = s.pos
		return scanMain
	case identStart:
		s.start = s.pos - 1
		return scanID
	default:
		return scanMain
	}
}

func scanID(s *scanner) stateFn {
	if s.peek() != identBracket {
		return scanMain
	}
	s.next()

	if !isNumber(s.peek()) {
		return scanMain
	}

	for isNumber(s.peek()) {
		s.next()
	}

	if s.peek() != identEnd {
		return scanMain
	}
	s.next()

	s.emit()
	return scanMain
}

func isNumber(r rune) bool {
	return r >= '0' && r <= '9'
}
