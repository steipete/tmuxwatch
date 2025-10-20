package tmux

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Session represents a tmux session with its windows populated.
type Session struct {
	ID        string
	Name      string
	Attached  bool
	CreatedAt time.Time
	Windows   []Window
}

// Window represents a tmux window and its panes.
type Window struct {
	ID       string
	Name     string
	Active   bool
	Session  string
	Index    int
	LastPane time.Time
	Panes    []Pane
}

// Pane represents a tmux pane.
type Pane struct {
	ID            string
	Title         string
	Active        bool
	Window        string
	Session       string
	CurrentCmd    string
	TTY           string
	LastActivity  time.Time
	CreatedAt     time.Time
	Width, Height int
}

// Snapshot contains the state of the tmux server.
type Snapshot struct {
	Sessions  []Session
	Timestamp time.Time
}

// paneContent represents a captured snapshot of a pane's buffer.
// Client interacts with the tmux binary.
type Client struct {
	bin string
}

// NewClient constructs a Client using the provided tmux binary path. If empty, LookPath("tmux") is used.
func NewClient(tmuxPath string) (*Client, error) {
	if tmuxPath == "" {
		var err error
		tmuxPath, err = exec.LookPath("tmux")
		if err != nil {
			return nil, fmt.Errorf("tmux not found in PATH: %w", err)
		}
	}
	return &Client{bin: tmuxPath}, nil
}

// Snapshot queries tmux and returns the current state of sessions/windows/panes.
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
		w := windows[i]
		// capture loop variable
		windowMap[w.ID] = &windows[i]
	}

	sessionMap := make(map[string]*Session, len(sessions))
	for i := range sessions {
		s := sessions[i]
		sessionMap[s.ID] = &sessions[i]
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

func (c *Client) listSessions(ctx context.Context) ([]Session, error) {
	cmd := exec.CommandContext(ctx, c.bin, "list-sessions", "-F", "#{session_id}\t#{session_name}\t#{session_attached}\t#{session_created}")
	out, err := cmd.Output()
	if err != nil {
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
		if len(fields) < 4 {
			return nil, fmt.Errorf("list-sessions: malformed line %q", line)
		}
		attached := fields[2] == "1"
		createdUnix, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid session_created %q: %w", fields[3], err)
		}
		session := Session{
			ID:        fields[0],
			Name:      fields[1],
			Attached:  attached,
			CreatedAt: time.Unix(createdUnix, 0),
		}
		sessions = append(sessions, session)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (c *Client) listWindows(ctx context.Context) ([]Window, error) {
	cmd := exec.CommandContext(ctx, c.bin, "list-windows", "-a", "-F", "#{session_id}\t#{window_id}\t#{window_index}\t#{window_name}\t#{window_active}\t#{window_last_flag}")
	out, err := cmd.Output()
	if err != nil {
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

func (c *Client) listPanes(ctx context.Context) ([]Pane, error) {
	format := strings.Join([]string{
		"#{session_id}",
		"#{window_id}",
		"#{pane_id}",
		"#{pane_active}",
		"#{pane_pid}",
		"#{pane_current_command}",
		"#{pane_title}",
		"#{pane_last_activity}",
		"#{pane_created}",
		"#{pane_width}",
		"#{pane_height}",
		"#{pane_tty}",
	}, "\t")

	cmd := exec.CommandContext(ctx, c.bin, "list-panes", "-a", "-F", format)
	out, err := cmd.Output()
	if err != nil {
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
		if len(fields) < 12 {
			return nil, fmt.Errorf("list-panes: malformed line %q", line)
		}
		active := fields[3] == "1"
		lastActivity, err := parseUnix(fields[7])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_last_activity %q: %w", fields[7], err)
		}
		created, err := parseUnix(fields[8])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_created %q: %w", fields[8], err)
		}
		width, err := strconv.Atoi(fields[9])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_width %q: %w", fields[9], err)
		}
		height, err := strconv.Atoi(fields[10])
		if err != nil {
			return nil, fmt.Errorf("invalid pane_height %q: %w", fields[10], err)
		}
		pane := Pane{
			Session:      fields[0],
			Window:       fields[1],
			ID:           fields[2],
			Active:       active,
			CurrentCmd:   fields[5],
			Title:        fields[6],
			LastActivity: lastActivity,
			CreatedAt:    created,
			Width:        width,
			Height:       height,
			TTY:          fields[11],
		}
		panes = append(panes, pane)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return panes, nil
}

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

// TitleOrCmd returns the best available label for a pane.
func (p Pane) TitleOrCmd() string {
	title := strings.TrimSpace(p.Title)
	if title != "" {
		return title
	}
	cmd := strings.TrimSpace(p.CurrentCmd)
	if cmd != "" {
		return cmd
	}
	return "pane"
}

// CapturePane retrieves the latest contents of a pane. lines determines how far back to capture.
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
