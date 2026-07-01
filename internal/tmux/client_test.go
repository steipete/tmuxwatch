package tmux

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func TestNewClientMissingTmux(t *testing.T) {
	t.Parallel()

	_, err := newClient("", func(string) (string, error) {
		return "", exec.ErrNotFound
	})
	if err == nil {
		t.Fatal("expected error when tmux is missing")
	}
	if !strings.Contains(err.Error(), "install tmux") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSessionsNoServer(t *testing.T) {
	t.Parallel()

	c := &Client{bin: "tmux", run: func(context.Context, string, ...string) ([]byte, error) {
		return nil, &exec.ExitError{Stderr: []byte("failed to connect to server")}
	}}

	sessions, err := c.listSessions(context.Background())
	if err != nil {
		t.Fatalf("listSessions returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions, got %d", len(sessions))
	}
}

func TestListWindowsNoServer(t *testing.T) {
	t.Parallel()

	c := &Client{bin: "tmux", run: func(context.Context, string, ...string) ([]byte, error) {
		return nil, &exec.ExitError{Stderr: []byte("no server running")}
	}}

	windows, err := c.listWindows(context.Background())
	if err != nil {
		t.Fatalf("listWindows returned error: %v", err)
	}
	if len(windows) != 0 {
		t.Fatalf("expected no windows, got %d", len(windows))
	}
}

func TestListPanesNoServer(t *testing.T) {
	t.Parallel()

	c := &Client{bin: "tmux", run: func(context.Context, string, ...string) ([]byte, error) {
		return nil, &exec.ExitError{Stderr: []byte("failed to connect to server")}
	}}

	panes, err := c.listPanes(context.Background())
	if err != nil {
		t.Fatalf("listPanes returned error: %v", err)
	}
	if len(panes) != 0 {
		t.Fatalf("expected no panes, got %d", len(panes))
	}
}
