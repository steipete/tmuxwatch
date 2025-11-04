// File options.go handles tmux option parsing helpers.
package tmux

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// PaneVariables returns user-defined (@-prefixed) tmux options scoped to a pane.
func (c *Client) PaneVariables(ctx context.Context, paneID string) (map[string]string, error) {
	if paneID == "" {
		return nil, fmt.Errorf("pane id cannot be empty")
	}
	cmd := exec.CommandContext(ctx, c.bin, "show-options", "-p", "-t", paneID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("show-options %s: %w", paneID, err)
	}

	vars := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		name, value, ok := parseOptionLine(scanner.Text())
		if !ok {
			continue
		}
		if strings.HasPrefix(name, "@") {
			vars[name] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return vars, nil
}

// parseOptionLine splits a tmux show-options line into name and value.
func parseOptionLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	parts := strings.SplitN(line, " ", 2)
	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "", "", false
	}
	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
	}
	return name, value, true
}
