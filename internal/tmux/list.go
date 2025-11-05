// Package tmux exposes helpers for translating tmux command output into rich
// Go structs the UI can consume.
package tmux

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var runCommand = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
	cmd := execCommand(ctx, bin, args...)
	if cmd == nil {
		return nil, fmt.Errorf("execCommand returned nil")
	}
	return cmd.Output()
}

// listSessions shells out to tmux to enumerate sessions and translate them
// into typed Session values.
func (c *Client) listSessions(ctx context.Context) ([]Session, error) {
	out, err := runCommand(ctx, c.bin, "list-sessions", "-F", "#{session_id}\t#{session_name}\t#{session_attached}\t#{session_created}\t#{session_activity}")
	if err != nil {
		if isNoServerError(err) {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("list-sessions: %w", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	sessions := []Session{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			return nil, fmt.Errorf("list-sessions: malformed line %q", line)
		}
		attached := fields[2] == "1"
		createdUnix, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid session_created %q: %w", fields[3], err)
		}
		lastActivity, err := parseUnix(fields[4])
		if err != nil {
			return nil, fmt.Errorf("invalid session_activity %q: %w", fields[4], err)
		}
		session := Session{
			ID:           fields[0],
			Name:         fields[1],
			Attached:     attached,
			CreatedAt:    time.Unix(createdUnix, 0),
			LastActivity: lastActivity,
		}
		sessions = append(sessions, session)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// listWindows retrieves every window in every session so we can later nest
// panes under them.
func (c *Client) listWindows(ctx context.Context) ([]Window, error) {
	out, err := runCommand(ctx, c.bin, "list-windows", "-a", "-F", "#{session_id}\t#{window_id}\t#{window_index}\t#{window_name}\t#{window_active}\t#{window_last_flag}")
	if err != nil {
		if isNoServerError(err) {
			return []Window{}, nil
		}
		return nil, fmt.Errorf("list-windows: %w", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	windows := []Window{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			return nil, fmt.Errorf("list-windows: malformed line %q", line)
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid window_index %q: %w", fields[2], err)
		}
		active := fields[4] == "1"
		window := Window{
			Session: fields[0],
			ID:      fields[1],
			Index:   index,
			Name:    fields[3],
			Active:  active,
		}
		windows = append(windows, window)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return windows, nil
}

// listPanes captures metadata for every pane so we can join them to windows
// and sessions.
func (c *Client) listPanes(ctx context.Context) ([]Pane, error) {
	format := strings.Join([]string{
		"#{session_id}",
		"#{window_id}",
		"#{pane_id}",
		"#{pane_active}",
		"#{pane_current_command}",
		"#{pane_title}",
		"#{pane_last_activity}",
		"#{pane_created}",
		"#{pane_width}",
		"#{pane_height}",
		"#{pane_tty}",
		"#{pane_dead}",
		"#{pane_dead_status}",
	}, "\t")

	out, err := runCommand(ctx, c.bin, "list-panes", "-a", "-F", format)
	if err != nil {
		if isNoServerError(err) {
			return []Pane{}, nil
		}
		return nil, fmt.Errorf("list-panes: %w", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	panes := []Pane{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 13 {
			return nil, fmt.Errorf("list-panes: malformed line %q", line)
		}
		active := fields[3] == "1"
		lastActivity, err := parseUnix(fields[6])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_last_activity %q: %w", fields[6], err)
		}
		created, err := parseUnix(fields[7])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_created %q: %w", fields[7], err)
		}
		width, err := strconv.Atoi(fields[8])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_width %q: %w", fields[8], err)
		}
		height, err := strconv.Atoi(fields[9])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_height %q: %w", fields[9], err)
		}
		pane := Pane{
			Session:      fields[0],
			Window:       fields[1],
			ID:           fields[2],
			Active:       active,
			CurrentCmd:   fields[4],
			Title:        fields[5],
			LastActivity: lastActivity,
			CreatedAt:    created,
			Width:        width,
			Height:       height,
			TTY:          fields[10],
			Dead:         fields[11] == "1",
		}
		if status := strings.TrimSpace(fields[12]); status != "" {
			if v, err := strconv.Atoi(status); err == nil {
				pane.DeadStatus = v
			}
		}
		panes = append(panes, pane)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return panes, nil
}

// parseUnix converts tmux's unix timestamp fields into a time value.
func parseUnix(v string) (time.Time, error) {
	if strings.TrimSpace(v) == "" {
		return time.Time{}, nil
	}
	iv, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(iv, 0), nil
}
