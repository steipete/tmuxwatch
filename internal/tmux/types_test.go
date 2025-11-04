package tmux

import "testing"

func TestPaneTitleOrCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pane Pane
		want string
	}{
		{
			name: "prefers title",
			pane: Pane{Title: "  logs  ", CurrentCmd: "bash"},
			want: "logs",
		},
		{
			name: "falls back to command",
			pane: Pane{CurrentCmd: "  go  "},
			want: "go",
		},
		{
			name: "defaults to pane",
			pane: Pane{},
			want: "pane",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pane.TitleOrCmd(); got != tt.want {
				t.Fatalf("TitleOrCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPaneStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pane Pane
		want string
	}{
		{name: "running", pane: Pane{Dead: false}, want: "running"},
		{name: "exit 0", pane: Pane{Dead: true, DeadStatus: 0}, want: "exit 0"},
		{name: "exit code", pane: Pane{Dead: true, DeadStatus: 3}, want: "exit 3"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pane.StatusString(); got != tt.want {
				t.Fatalf("StatusString() = %q, want %q", got, tt.want)
			}
		})
	}
}
