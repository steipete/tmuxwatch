# tmuxwatch

`tmuxwatch` is a Go-powered terminal user interface that keeps an eye on your tmux server and gives you a live dashboard of sessions, windows, and panes. New activity shows up automatically, fresh sessions pop into view with their recent output, and the layout adapts so you can focus on the streams that matter.

## What It Does

- Polls `tmux list-*` commands and renders a live dashboard so you can glance at everything that is running without detaching or memorising `tmux` incantations.
- Captures the active pane of every session and displays the latest output in scrollable cards that automatically follow new logs unless you scroll away manually.
- Adapts the preview grid to however many sessions you have, giving each one a fair slice of terminal real estate.
- Includes a `--dump` CLI flag so scripts (or curious humans) can print the current tmux state as JSON without launching the TUI.
- Includes a `--dump` CLI flag so scripts (or curious humans) can print the current tmux state as JSON without launching the TUI.

Use tmuxwatch when you want the situational awareness of a monitoring dashboard but prefer to stay inside a terminal workflow.

## Why

- Quickly see which tmux sessions are active without memorising `tmux list-…` incantations.
- Maintain a focused workspace by spotting noisy panes instantly and filtering by session, window, or command.
- Take advantage of Charmbracelet's Bubble Tea stack for a polished, keyboard-driven TUI.

## Architecture

- **CLI entrypoint** (`cmd/tmuxwatch/main.go`): parses flags, sets up the tmux client, and launches the Bubble Tea program.
- **Tmux adapter** (`internal/tmux`): thin wrapper around the `tmux` binary that shells out to `list-sessions`, `list-windows`, and `list-panes`, parsing their structured output into Go structs. Periodically polls using a ticker (default 1 s) and emits diff events to the UI.
- **TUI model** (`internal/ui`): Bubble Tea model that renders one scrollable viewport per tmux session. Each card shows the active window/pane, auto-refreshes its capture-pane output, and shares the available height evenly across sessions. A search bar filters sessions by title, window name, or pane command in real time.

## Key Bindings

- `/` or `ctrl+f`: open the search bar to filter sessions/windows/panes
- `esc`: close the search bar (clears focus) or abort the filter
- `q` / `ctrl+c`: quit
- Scroll inside a session card with standard viewport bindings (`ctrl+d`, `ctrl+u`, `ctrl+f`, `ctrl+b`, `g`, `G`)

## Requirements

- Go 1.25+
- tmux 3.1+ available in `$PATH`

## Development

```bash
go run ./cmd/tmuxwatch
```

`go run` compiles the latest sources before executing, so you always see your current edits—no extra build step required.

For a quick inspection of the tmux state without entering the TUI, use the debug dump mode:

```bash
go run ./cmd/tmuxwatch --dump
```

### Tooling

Install and use formatter/linter binaries that track the versions pinned in `tools/tools.go`:

```bash
go install mvdan.cc/gofumpt@v0.9.1
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

make fmt   # gofumpt -w .
make lint  # golangci-lint run
make check # runs both
```

Useful targets that will be added later:

- `go test ./...`
- `golangci-lint run`

## Release Checklist

- Update `CHANGELOG.md`
- Bump version in future goreleaser config
- Tag release (`git tag vX.Y.Z && git push --tags`)
