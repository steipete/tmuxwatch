// File commands.go defines Bubble Tea command constructors for tmuxwatch.
package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/steipete/tmuxwatch/internal/tmux"
)

// fetchSnapshotCmd captures the current tmux snapshot or returns an error
// message when it fails.
func fetchSnapshotCmd(client *tmux.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		snap, err := client.Snapshot(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return snapshotMsg{snapshot: snap}
	}
}

// scheduleTick creates a periodic timer used to refresh tmux snapshots.
func scheduleTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// fetchPaneContentCmd grabs the latest pane output for preview rendering.
func fetchPaneContentCmd(client *tmux.Client, sessionID, paneID string, lines int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		text, err := client.CapturePane(ctx, paneID, lines)
		return paneContentMsg{sessionID: sessionID, paneID: paneID, text: text, err: err}
	}
}

// fetchPaneVarsCmd loads user-defined tmux variables for the provided pane.
func fetchPaneVarsCmd(client *tmux.Client, sessionID, paneID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		vars, err := client.PaneVariables(ctx, paneID)
		return paneVarsMsg{sessionID: sessionID, paneID: paneID, vars: vars, err: err}
	}
}

// sendKeysCmd forwards keystrokes to a tmux pane within a context deadline.
func sendKeysCmd(client *tmux.Client, paneID string, keys ...string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.SendKeys(ctx, paneID, keys...); err != nil {
			return errMsg{err: err}
		}
		return nil
	}
}

// killSessionCmd terminates a tmux session and triggers a refresh.
func killSessionCmd(client *tmux.Client, sessionID string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        if err := client.KillSession(ctx, sessionID); err != nil {
            return errMsg{err: err}
        }
        return killSessionMsg{sessionID: sessionID}
    }
}
