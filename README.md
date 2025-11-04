# tmuxwatch

tmuxwatch is a terminal dashboard that keeps eyes on every tmux session, window, and pane in real time. It polls tmux, captures active pane output, and renders a Bubble Tea UI so you can stay in-monitoring mode without leaving the terminal.

## Features
- **Live overview**: Sessions appear as cards with continuously refreshed pane output and status pulses when logs change.
- **Keyboard-first controls**: Filter with `/`, navigate cards like a log viewer, and forward keystrokes directly into the underlying pane.
- **Mouse support**: Click to focus, scroll to browse history, or close a session via the inline `[x]`.
- **Automation friendly**: `--dump` flag prints the current tmux topology as JSON for scripts or quick debugging.

## Quick Start
```sh
tmux new-session -d -s watch 'go run ./cmd/tmuxwatch'
tmux attach -t watch
```
Press `q` (or double `ctrl+c`) to exit. Prefer running tmuxwatch in its own tmux session to keep the UI isolated from your workspaces.

## CLI Flags
- `--interval <duration>`: Poll frequency (default `1s`).
- `--tmux <path>`: Path to the tmux binary; defaults to `$PATH` lookup.
- `--dump`: Print the current snapshot as indented JSON and exit.
- `--version`: Display the build version and exit.

## Architecture Overview
- `cmd/tmuxwatch/`: CLI entry point wiring flags, versioning, and Bubble Tea program setup.
- `internal/tmux/`: Shells out to tmux, parses `list-*` output, and exposes reusable snapshot types plus capture/send helpers.
- `internal/ui/`: Modular Bubble Tea model split into model/update/handlers/layout/render/utility files and matching unit tests.
- `docs/`: Product and UX backlogs; `AGENTS.md` offers contributor directives.

## Development
```sh
go build ./...
go test ./...
make fmt && make lint
tmux new-session -d -s watch 'go run ./cmd/tmuxwatch --dump'; tmux kill-session -t watch
```
Always run the application inside tmux; never invoke `go run` directly from a bare shell when verifying behaviour.

## License
Released under the [MIT License](./LICENSE).
