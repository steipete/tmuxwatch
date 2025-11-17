<shared>
# AGENTS.md

Shared guardrails distilled from the various `~/Projects/*/AGENTS.md` files (state as of **November 15, 2025**). This document highlights the rules that show up again and again; still read the repo-local instructions before making changes.

## Codex Global Instructions
- Keep the system-wide Codex guidance at `~/.codex/AGENTS.md` (the Codex home; override via `CODEX_HOME` if needed) so every task inherits these rules by default.

## General Guardrails

### Intake & Scoping
- Open the local agent instructions plus any `docs:list` summaries at the start of every session. Re-run those helpers whenever you suspect the docs may have changed.
- Review any referenced tmux panes, CI logs, or failing command transcripts so you understand the most recent context before writing code.

### Tooling & Command Wrappers
- Use the command wrappers provided by the workspace (`./runner …`, `scripts/committer`, `pnpm mcp:*`, etc.). Skip them only for trivial read-only shell commands if that’s explicitly allowed.
- Stick to the package manager and runtime mandated by the repo (pnpm-only, bun-only, swift-only, go-only, etc.). Never swap in alternatives without approval.
- When editing shared guardrail scripts (runners, committer helpers, browser tools, etc.), mirror the same change back into the `agent-scripts` folder so the canonical copy stays current.
- Ask the user before adding dependencies, changing build tooling, or altering project-wide configuration.
- Keep the project’s `AGENTS.md` `<tools>
# TOOLS

Edit guidance: keep the actual tool list inside the `<tools></tools>` block below so downstream AGENTS syncs can copy the block contents verbatim (without wrapping twice).

<tools>
- `runner`: Bash shim that routes every command through Bun guardrails (timeouts, git policy, safe deletes).
- `git` / `bin/git`: Git shim that forces git through the guardrails; use `./git --help` to inspect.
- `scripts/committer`: Stages the files you list and creates the commit safely.
- `scripts/docs-list.ts`: Walks `docs/`, enforces front-matter, prints summaries; run `tsx scripts/docs-list.ts`.
- `scripts/browser-tools.ts`: Chrome helper for remote control/screenshot/eval; run `ts-node scripts/browser-tools.ts --help`.
- `scripts/runner.ts`: Bun implementation backing `runner`; run `bun scripts/runner.ts --help`.
- `bin/sleep`: Sleep shim that enforces the 30s ceiling; run `bin/sleep --help`.
- `xcp`: Xcode project/workspace helper; run `xcp --help`.
- `oracle`: CLI to bundle prompt + files for another AI; run `npx -y @steipete/oracle --help`.
- `mcporter`: MCP launcher for any registered MCP server; run `npx mcporter`.
- `iterm`: Full TTY terminal via MCP; run `npx mcporter iterm`.
- `firecrawl`: MCP-powered site fetcher to Markdown; run `npx mcporter firecrawl`.
- `XcodeBuildMCP`: MCP wrapper around Xcode tooling; run `npx mcporter XcodeBuildMCP`.
- `gh`: GitHub CLI for PRs, CI logs, releases, repo queries; run `gh help`.
</tools>

</tools>
` block in sync with the full tool list from `TOOLS.md` so downstream repos get the latest tool descriptions.

### tmux & Long Tasks
- Run any command that could hang (tests, servers, log streams, browser automation) inside tmux using the repository’s preferred entry point.
- Do not wrap tmux commands in infinite polling loops. Run the job, sleep briefly (≤30 s), capture output, and surface status at least once per minute.
- Document which sessions you create and clean them up when they are no longer needed unless the workflow explicitly calls for persistent watchers.

### Build, Test & Verification
- Before handing off work, run the full “green gate” for that repo (lint, type-check, tests, doc scripts, etc.). Follow the same command set humans run—no ad-hoc shortcuts.
- Leave existing watchers running unless the owner tells you to stop them; keep their tmux panes healthy if you started them.
- Treat every bug fix as a chance to add or extend automated tests that prove the behavior.

### Code Quality & Naming
- Refactor in place. Never create duplicate files with suffixes such as “V2”, “New”, or “Fixed”; update the canonical file and remove obsolete paths entirely.
- Favor strict typing: avoid `any`, untyped dictionaries, or generic type erasure unless absolutely required. Prefer concrete structs/enums and mark public concurrency surfaces appropriately.
- Keep files at a manageable size. When a file grows unwieldy, extract helpers or new modules instead of letting it bloat.
- Match the repo’s established style (commit conventions, formatting tools, component patterns, etc.) by studying existing code before introducing new patterns.

### Git, Commits & Releases
- Invoke git through the provided wrappers, especially for status, diffs, and commits. Only commit or push when the user asks you to do so.
- Follow the documented release or deployment checklists instead of inventing new steps.
- Do not delete or rename unfamiliar files without double-checking with the user or the repo instructions.

### Documentation & Knowledge Capture
- Update existing docs whenever your change affects them, including front-matter metadata if the repo’s `docs:list` tooling depends on it.
- Only create new documentation when the user or local instructions explicitly request it; otherwise, edit the canonical file in place.
- When you uncover a reproducible tooling or CI issue, record the repro steps and workaround in the designated troubleshooting doc for that repo.

### Troubleshooting & Observability
- Design workflows so they are observable without constant babysitting: use tmux panes, CI logs, log-tail scripts, MCP/browser helpers, and similar tooling to surface progress.
- If you get stuck, consult external references (web search, official docs, Stack Overflow, etc.) before escalating, and record any insights you find for the next agent.
- Keep any polling or progress loops bounded to protect hang detectors and make it obvious when something stalls.

### Stack-Specific Reminders
- Start background builders or watchers using the automation provided by the repo (daemon scripts, tmux-based dev servers, etc.) instead of running binaries directly.
- Use the official CLI wrappers for browser automation, screenshotting, or MCP interactions rather than crafting new ad-hoc scripts.
- Respect each workspace’s testing cadence (e.g., always running the main `check` script after edits, never launching forbidden dev servers, keeping replies concise when requested).

## Swift Projects
- Kick off the workspace’s build daemon or helper before running any Swift CLI or app; rely on the provided wrapper to rebuild targets automatically instead of launching stale binaries.
- Validate changes with `swift build` and the relevant filtered test suites, documenting any compiler crashes and rewriting problematic constructs immediately so the suite can keep running.
- Keep concurrency annotations (`Sendable`, actors, structured tasks) accurate and prefer static imports over dynamic runtime lookups that break ahead-of-time compilation.
- Avoid editing derived artifacts or generated bundles directly—adjust the sources and let the build pipeline regenerate outputs.
- When encountering toolchain instability, capture the repro steps in the designated troubleshooting doc and note any required cache cleans (DerivedData, SwiftPM caches) you perform.

## TypeScript Projects
- Use the package manager declared by the workspace (often `pnpm` or `bun`) and run every command through the same wrapper humans use; do not substitute `npm`/`yarn` or bypass the runner.
- Start each session by running the repo’s doc-index script (commonly a `docs:list` helper), then keep required watchers (`lint:watch`, `test:watch`, dev servers) running inside tmux unless told otherwise.
- Treat `lint`, `typecheck`, and `test` commands (e.g., `pnpm run check`, `bun run typecheck`) as mandatory gates before handing off work; surface any failures with their exact command output.
- Maintain strict typing—avoid `any`, prefer utility helpers already provided by the repo, and keep shared guardrail scripts (runner, committer, browser helpers) consistent by syncing back to `agent-scripts` when they change.
- When editing UI code, follow the established component patterns (Tailwind via helper utilities, TanStack Query for data flow, etc.) and keep files under the preferred size limit by extracting helpers proactively.

Keep this master file up to date as you notice new rules that recur across repositories, and reflect those updates back into every workspace’s local guardrail documents.

</shared>

<tools>
# TOOLS

Edit guidance: keep the actual tool list inside the `<tools></tools>` block below so downstream AGENTS syncs can copy the block contents verbatim (without wrapping twice).

<tools>
- `runner`: Bash shim that routes every command through Bun guardrails (timeouts, git policy, safe deletes).
- `git` / `bin/git`: Git shim that forces git through the guardrails; use `./git --help` to inspect.
- `scripts/committer`: Stages the files you list and creates the commit safely.
- `scripts/docs-list.ts`: Walks `docs/`, enforces front-matter, prints summaries; run `tsx scripts/docs-list.ts`.
- `scripts/browser-tools.ts`: Chrome helper for remote control/screenshot/eval; run `ts-node scripts/browser-tools.ts --help`.
- `scripts/runner.ts`: Bun implementation backing `runner`; run `bun scripts/runner.ts --help`.
- `bin/sleep`: Sleep shim that enforces the 30s ceiling; run `bin/sleep --help`.
- `xcp`: Xcode project/workspace helper; run `xcp --help`.
- `oracle`: CLI to bundle prompt + files for another AI; run `npx -y @steipete/oracle --help`.
- `mcporter`: MCP launcher for any registered MCP server; run `npx mcporter`.
- `iterm`: Full TTY terminal via MCP; run `npx mcporter iterm`.
- `firecrawl`: MCP-powered site fetcher to Markdown; run `npx mcporter firecrawl`.
- `XcodeBuildMCP`: MCP wrapper around Xcode tooling; run `npx mcporter XcodeBuildMCP`.
- `gh`: GitHub CLI for PRs, CI logs, releases, repo queries; run `gh help`.
</tools>

</tools>

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
- Use [Conventional Commits v1.0](https://www.conventionalcommits.org/en/v1.0.0/) **without exception**. Allowed types: `feat|fix|refactor|build|ci|chore|docs|style|perf|test`. Variants such as `feat(ui): ...` or `chore!: ...` are welcome. Example messages: `feat: prevent racing of requests`, `chore!: drop support for iOS 16`, `feat(api): add basic telemetry`.
- Write the description in the imperative mood and bundle only related changes per commit.
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
