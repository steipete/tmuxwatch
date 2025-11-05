// Package tmux handles communication with the tmux binary for snapshotting and
// interacting with running panes.
package tmux

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"time"
)

var (
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, name, args...)
	}
	lookPath = exec.LookPath
)

// Client wraps a tmux binary path and exposes high-level snapshot helpers.
type Client struct {
	bin string
}

// NewClient constructs a Client using the provided tmux binary. When tmuxPath
// is empty, LookPath("tmux") is used so the system PATH controls discovery.
func NewClient(tmuxPath string) (*Client, error) {
	if tmuxPath == "" {
		var err error
		tmuxPath, err = lookPath("tmux")
		if err != nil {
			return nil, fmt.Errorf("tmux not found in PATH (install tmux >=3.1): %w", err)
		}
	}
	return &Client{bin: tmuxPath}, nil
}

// Snapshot queries tmux for sessions, windows, and panes and returns a unified
// structure ready for presentation.
func (c *Client) Snapshot(ctx context.Context) (Snapshot, error) {
	sessions, err := c.listSessions(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	windows, err := c.listWindows(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	panes, err := c.listPanes(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	windowMap := make(map[string]*Window, len(windows))
	for i := range windows {
		windowMap[windows[i].ID] = &windows[i]
	}

	sessionMap := make(map[string]*Session, len(sessions))
	for i := range sessions {
		sessionMap[sessions[i].ID] = &sessions[i]
	}

	for _, pane := range panes {
		if win, ok := windowMap[pane.Window]; ok {
			win.Panes = append(win.Panes, pane)
			if pane.LastActivity.After(win.LastPane) {
				win.LastPane = pane.LastActivity
			}
		}
	}

	for _, window := range windows {
		if session, ok := sessionMap[window.Session]; ok {
			session.Windows = append(session.Windows, window)
		}
	}

	return Snapshot{
		Sessions:  slices.Clone(sessions),
		Timestamp: time.Now(),
	}, nil
}

// CapturePane retrieves lines of output from a tmux pane for preview rendering.
func (c *Client) CapturePane(ctx context.Context, paneID string, lines int) (string, error) {
	if paneID == "" {
		return "", fmt.Errorf("pane id cannot be empty")
	}
	if lines <= 0 {
		lines = 200
	}
	start := fmt.Sprintf("-%d", lines)
	cmd := exec.CommandContext(ctx, c.bin, "capture-pane", "-p", "-J", "-t", paneID, "-S", start)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("capture-pane %s: %w", paneID, err)
	}
	return string(out), nil
}

// SendKeys forwards key sequences to a tmux pane so the user can interact with
// it through tmuxwatch.
func (c *Client) SendKeys(ctx context.Context, paneID string, keys ...string) error {
	if paneID == "" {
		return fmt.Errorf("pane id cannot be empty")
	}
	if len(keys) == 0 {
		return nil
	}
	args := append([]string{"send-keys", "-t", paneID}, keys...)
	cmd := exec.CommandContext(ctx, c.bin, args...)
	return cmd.Run()
}

// KillSession terminates a tmux session by id.
func (c *Client) KillSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session id cannot be empty")
	}
	cmd := exec.CommandContext(ctx, c.bin, "kill-session", "-t", sessionID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kill-session %s: %w", sessionID, err)
	}
	return nil
}
