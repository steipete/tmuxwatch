# tmuxwatch 👀

`tmuxwatch` is a Charmbracelet-powered dashboard that keeps eyes on every tmux session, window, and pane without ever leaving the terminal.

## Highlights
- **Live tmux snapshot**: Polls `list-sessions`, `list-windows`, and `list-panes`, stitches the hierarchy together, and shows the latest capture-pane output per session.
- **Tab-aware layout**: The strip lists the grid plus every visible tmux session; click or `shift+left/right` to jump tabs, `ctrl+m` toggles full-screen, and `esc` returns to the grid.
- **Keyboard & mouse aware**: `/` to search, arrow/PageUp/PageDown to scroll, collapse cards with `z`/`Z`, maximise via `ctrl+m` or the `[^]` control, `X` to kill a focused stale session, `ctrl+X` to clean *all* stale sessions, and mouse clicks/scrolls to focus, collapse, close cards, or switch tabs.
- **Command palette (`ctrl+P`)**: Run actions (refresh, show hidden, clean stale) from a centered overlay.
- **Automation friendly**: `--dump` prints the current tmux topology as JSON for scripts or debugging.

## Install & Run
```sh
# Homebrew (recommended)
brew tap steipete/tap
brew install tmuxwatch
tmuxwatch --version  # should print tmuxwatch 0.9

# Updating later
brew update
brew upgrade tmuxwatch

# Or install directly with Go tooling
go install github.com/steipete/tmuxwatch/cmd/tmuxwatch@latest

# best practice: spawn inside tmux so key bindings work as expected
tmux new-session -d -s watch './gorunfresh --trace-mouse'
tmux attach -t watch

# prefer to run outside tmux?
./gorunfresh --dump   # runs directly unless TMUXWATCH_FORCE_TMUX=1 is set
./gorunfresh --watch  # starts poltergeist + polter --watch to auto-restart after successful builds
./scripts/run-hot.sh  # single command that starts daemon in the background and launches tmuxwatch with hot reload
```
Press `q` (or double `ctrl+c`) to exit. Prefer running tmuxwatch in its own tmux session to keep the UI isolated from your workspaces. For local development you can substitute `./gorunfresh --debug-click 30,10 --trace-mouse` inside the session to replay a mouse event while inspecting BubbleZone logs.

## CLI Flags
- `--interval <duration>`: tmux poll frequency (default `1s`).
- `--tmux <path>`: tmux binary to execute (defaults to `$PATH`).
- `--dump`: emit the current snapshot as indented JSON and exit.
- `--version`: print the build/version string.

## Keyboard & Mouse Cheat Sheet
```
/ or ctrl+f        open search; type to filter sessions/windows/panes
esc                clear search, close palette, or leave detail view
shift+left/right   switch tabs
H                  show hidden sessions
X                  kill the focused stale session
ctrl+X             kill every stale session
ctrl+P             open/close the command palette
ctrl+m             maximise/restore the focused session
z / Z              collapse focused session / expand all sessions
q / ctrl+c         quit (double ctrl+c quits even if pane is alive)
mouse              click `[^]/[v]` to maximise/restore, `[-]/[+]` to collapse/expand, `[x]` to hide; scroll to browse logs
```

## Architecture
- `cmd/tmuxwatch/`: CLI entry point, flag parsing, Bubble Tea program setup.
- `internal/tmux/`: thin wrapper over the tmux binary (snapshot capture, capture-pane, send-keys, kill-session, option queries).
- `internal/ui/`: Bubble Tea model split into focused files (`model`, `update`, `handlers`, `cards`, `status`, `palette`, `overlay`, etc.).
- `docs/`: contributor docs (`AGENTS.md`, `idiomatic-go.md`).

The UI intentionally avoids third-party “magic”; it leans on Bubble Tea + Lip Gloss primitives so behaviour is explicit.

## Development Workflow
Use the pnpm scripts to mirror the Go tooling:
```sh
pnpm format  # gofumpt -w .
pnpm lint    # golangci-lint run
pnpm test    # go test ./...
pnpm build   # go build ./cmd/tmuxwatch
pnpm start   # runs ./gorunfresh (prefer inside tmux)
```

You can still call the Go targets directly:
```sh
# format + lint
make fmt
make lint

# run tests (includes table-driven/unit tests in internal/ui and internal/tmux)
go test ./...

# fresh rebuild & run inside tmux (guards against outside usage)
./gorunfresh --dump  # validate tmux JSON snapshot; respects TMUXWATCH_FORCE_TMUX

go run ./cmd/tmuxwatch --dump  # (inside tmux) validate tmux JSON snapshot for debugging
```
Guidelines live in `docs/idiomatic-go.md`; treat it as required reading. Key points:
- Always run the app inside tmux; tests that touch tmux spawn/destroy their own sessions.
- Run `gofumpt`, `golangci-lint`, and `govulncheck` before opening a PR.
- Table-driven tests go beside their packages (`*_test.go`); fixtures live under `testdata/`.
- The `gorunfresh` helper clears cache and re-runs the app but refuses to execute unless you launch it from within tmux.

## Roadmap (short list)
- Theming + palette customization (Catppuccin/Dracula).
- Configurable capture depth & poll interval via config file.
- Pane interaction history and saved layouts.

## License
Released under the [MIT License](./LICENSE).
