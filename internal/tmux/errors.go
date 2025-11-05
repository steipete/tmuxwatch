package tmux

import (
	"errors"
	"os/exec"
	"strings"
)

var noServerHints = []string{
	"failed to connect to server",
	"no server running",
	"error connecting to",
}

func isNoServerError(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		lower := strings.ToLower(string(exitErr.Stderr))
		for _, hint := range noServerHints {
			if strings.Contains(lower, hint) {
				return true
			}
		}
	}
	lower := strings.ToLower(err.Error())
	for _, hint := range noServerHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}
