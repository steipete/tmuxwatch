# Repository Guidelines

## Project Structure & Module Organization
- `cmd/tmuxwatch/` holds the CLI entry point and flag handling.
- `internal/tmux/` wraps the tmux binary, exposes snapshot types, and contains parser tests.
- `internal/ui/` owns the Bubble Tea model; sub-files group model, update loop, handlers, layout, commands, rendering, and utilities alongside targeted tests.
- `docs/` stores design notes; `tools/` pins developer tooling; `Makefile` defines fmt/lint shortcuts.
- Tests live beside implementation files (`*_test.go`) for quick discovery.

## Build, Test, and Development Commands
- `go build ./...` — compile every package to surface type errors early.
- `go test ./...` — execute table-driven unit tests; required before commits.
- `tmux new-session -d -s watch 'go run ./cmd/tmuxwatch --dump'` — validate snapshot output inside tmux; kill with `tmux kill-session -t watch`.
- `make fmt` / `make lint` — run `gofumpt` and `golangci-lint` using the repository’s pinned toolchain.

## Coding Style & Naming Conventions
- Use Go defaults: tabs for indentation, exported identifiers with descriptive CamelCase names, private helpers in lowerCamelCase.
- Always format with `gofumpt` (already vendored via `Makefile`; run before commits).
- Keep files focused; split large components into purpose-specific siblings as done in `internal/ui/`.

## Testing Guidelines
- Prefer table-driven tests mirroring the patterns in `internal/tmux/types_test.go` and `internal/ui/util_test.go`.
- Name tests `Test<Subject>` and mark long-running cases with `t.Skip`.
- Run `go test ./...` locally and ensure new behaviour is covered; document edge cases in the test table comments when non-obvious.

## Commit & Pull Request Guidelines
- Write commits in the imperative mood (`Add Bubble Tea layout helpers`), bundling related changes only.
- Reference relevant issues in commit bodies or PR descriptions.
- PRs should summarize user-facing impact, list validation commands (`go test ./...`, `go run ./cmd/tmuxwatch --dump`), and attach screenshots/gifs if UI output changes.

## Environment & Tooling Notes
- Development requires tmux ≥3.1 on PATH; all tmuxwatch runs must occur inside a tmux session.
- Keep pinned tool versions (`go.mod`, `Makefile`) aligned after dependency bumps; finish with `go mod tidy`.

## Agent Directives & Learnings
- Never launch the TUI directly from the shell; wrap checks in tmux (`tmux new-session ... go run ./cmd/tmuxwatch`).
- Prefer `gofumpt`+`golangci-lint` for consistency; avoid ad-hoc formatters.
- When scripting tmux interactions for tests, remember to clean up (`tmux kill-session`) to leave developer sessions untouched.
- Treat `docs/idiomatic-go.md` as required reading before contributing; follow its 2025 idioms for style, tooling, and testing norms.
- Use `./gorunfresh` for rebuild-and-run workflows; it supports `--debug-click`/`--trace-mouse` for BubbleZone debugging and honours `TMUXWATCH_FORCE_TMUX=1` if you want to require a tmux session.
- BubbleZone underpins mouse hit-testing (cards, controls, and tab titles); prefer zone-aware helpers/tests instead of manual geometry math.
- Follow `RELEASE.md` when tagging new versions and updating the Homebrew tap in `~/Projects/homebrew-tap`.
