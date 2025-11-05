package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
)

func TestIsNoServerError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"exit without hint", &exec.ExitError{}, false},
		{"exit with failed message", &exec.ExitError{Stderr: []byte("failed to connect to server\n")}, true},
		{"exit with no server message", &exec.ExitError{Stderr: []byte("no server running on /tmp/tmux-1000/default")}, true},
		{"wrapped error", fmt.Errorf("wrapper: %w", &exec.ExitError{Stderr: []byte("failed to connect to server")}), true},
		{"plain error string", errors.New("no server running"), true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isNoServerError(tc.err); got != tc.want {
				t.Fatalf("isNoServerError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
