package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTmuxKeysFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		msg    tea.KeyMsg
		want   []string
		expect bool
	}{
		{name: "enter", msg: tea.KeyMsg{Type: tea.KeyEnter}, want: []string{"Enter"}, expect: true},
		{name: "space", msg: tea.KeyMsg{Type: tea.KeySpace}, want: []string{" "}, expect: true},
		{name: "alt runes rejected", msg: tea.KeyMsg{Type: tea.KeyRunes, Alt: true}, expect: false},
		{name: "runes ok", msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, want: []string{"a"}, expect: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := tmuxKeysFrom(tt.msg)
			if ok != tt.expect {
				t.Fatalf("tmuxKeysFrom ok = %v, want %v", ok, tt.expect)
			}
			if !tt.expect {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("tmuxKeysFrom len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("tmuxKeysFrom got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
