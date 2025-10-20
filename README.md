# tmuxwatch

`tmuxwatch` is a Go-powered terminal user interface that keeps an eye on your tmux server and gives you a live dashboard of sessions, windows, and panes. New activity shows up automatically, panes you do not care about can be hidden with a keystroke, and the layout adapts so you can focus on the streams that matter.

## Why

- Quickly see which tmux sessions are active without memorising `tmux list-…` incantations.
- Maintain a focused workspace by hiding noisy panes.
- Take advantage of Charmbracelet's Bubble Tea stack for a polished, keyboard-driven TUI.

## Architecture

- **CLI entrypoint** (`cmd/tmuxwatch/main.go`): parses flags, sets up the tmux client, and launches the Bubble Tea program.
- **Tmux adapter** (`internal/tmux`): thin wrapper around the `tmux` binary that shells out to `list-sessions`, `list-windows`, and `list-panes`, parsing their structured output into Go structs. Periodically polls using a ticker (default 1 s) and emits diff events to the UI.
- **TUI model** (`internal/ui`): Bubble Tea model with three panes:
  - left sidebar: list of sessions and windows, navigated with `j/k` or arrow keys;
  - main view: tab bar of panes for the selected window with live buffer previews captured via `tmux capture-pane`;
  - status/footer: key hints (navigation, reorder, hide) and last refresh time via Lip Gloss styling.
- **Hidden pane manager**: tracks pane IDs toggled via `h` (hide) and can be surfaced again with `H` (show all).

## Key Bindings

- `q` / `ctrl+c`: quit
- `j` / `down`, `k` / `up`: move selection in the sidebar
- `enter` / `l`: switch focus from sidebar to panes
- `tab` / `shift+tab`: toggle focus between sidebar and pane tabs
- `j` / `down`, `k` / `up`, `left` / `right`: switch between pane tabs
- `[` / `]`: reorder the focused pane tab
- `h`: hide the focused pane
- `H`: reveal all hidden panes
- `r`: force refresh from tmux
- `q` / `ctrl+c`: quit

## Requirements

- Go 1.25+
- tmux 3.1+ available in `$PATH`

## Development

```bash
go run ./cmd/tmuxwatch
```

For a quick inspection of the tmux state without entering the TUI, use the debug dump mode:

```bash
go run ./cmd/tmuxwatch --dump
```

Useful targets that will be added later:

- `go test ./...`
- `golangci-lint run`

## Release Checklist

- Update `CHANGELOG.md`
- Bump version in future goreleaser config
- Tag release (`git tag vX.Y.Z && git push --tags`)
