# tmuxwatch Spec

## Vision

Deliver a tmux companion TUI that gives immediate situational awareness across every session, window, and pane. Users should be able to:

- See live pane output in real time without manually iterating through `tmux` commands.
- Triage noisy panes by hiding or reordering them, so attention stays on high-signal streams.
- Move through their tmux estate entirely by keyboard (with optional mouse support later).
- Eventually snapshot and restore tmux workspaces to jump between projects quickly.

## Core User Stories

1. **Live Monitor**: “As a developer, I want to open tmuxwatch and instantly see the latest output of panes across all active sessions so I can detect failures or progress at a glance.”
2. **Focus Control**: “As a developer, I want to skip panes that aren’t relevant right now by hiding or reordering them, so my dashboard stays uncluttered.”
3. **Navigation**: “As a tmux power user, I want vim-style keyboard navigation between sessions, windows, and panes, to keep muscle memory intact.”
4. **Debug View**: “As an engineer diagnosing issues, I want a non-interactive mode that prints the current tmux structure as JSON so scripts and humans can inspect state.”
5. **Future Snapshot** *(planned)*: “As a multitasker, I want to save and restore full tmux layouts, including pane commands and paths, to switch projects effortlessly.”

## Current Architecture

- **`internal/tmux`**: Go wrapper around the tmux binary, providing structured snapshots plus capture-pane support.
- **`internal/ui`**: Bubble Tea model that renders adaptive session preview cards. Cards capture the active window/pane output, auto-scroll with new data, expose session metadata (last activity, exit status), forward keystrokes to panes, pulse briefly when output changes, and provide mouse affordances (focus, scroll, close).
- **`cmd/tmuxwatch`**: CLI entry point with flags (`--interval`, `--tmux`, `--dump`) and version reporting.
- **Documentation**: README highlights usage, features, and key bindings; changelog tracks notable updates; MIT license governs distribution.

## Inspiration & Competitive Scan

- **tmux-tui (Haskell)**: demonstrates stored session persistence, fuzzy search filters, and petname-based session creation. Highlights value of JSON snapshots, confirmation dialogs, and mouse-aware focus rings.
- **tmux_tui (Go)**: provides split-pane helpers, swap modes, filtering, and theming. Reinforces demand for:
  - Fast `list-panes` driven updates paired with `display-message` for current focus detection.
  - Inline preview panel using `capture-pane -ep`.
  - Swap workflows (mark source, accept destination) for windows/panes.
  - Themable UI via predefined palettes (e.g., Catppuccin, Dracula, Nord).
- tmuxwatch should blend best ideas: keep lightweight monitoring focus while adopting discoverable commands (swap, rename, filters) and optional theming.

## Roadmap

### Phase 1 — Monitoring Foundations *(done)*
- [x] Poll tmux sessions/windows/panes.
- [x] Render adaptive session preview cards that auto-scroll with live output.
- [x] Provide `--dump` JSON snapshot for debugging/automation.

### Phase 2 — UX Enhancements *(in progress)*
- [ ] Add `--capture-lines` flag and config file to control history depth.
- [x] Implement search/filter across sessions, windows, and panes.
- [x] Forward focus and keystrokes from tmuxwatch to live panes; add mouse support for focusing, scrolling, and closing cards.
- [ ] Surface more status metadata: pane last activity, command exit statuses, alerts.
- [ ] Introduce optional theme selection aligned with tmux_tui palettes (Dracula, Nord, Catppuccin, etc.).
- [ ] Provide swap workflows for panes/windows with visual feedback (mark source, confirm target).

### Phase 3 — Workspace Management *(future)*
- [ ] Introduce snapshot persistence (store sessions/windows/panes as JSON in `~/.config/tmuxwatch/`).
- [ ] Provide commands to instantiate stored sessions, similar to Haskell tmux-tui.
- [ ] Offer YAML/JSON schema for curated dashboards (e.g., always pin specific panes).
- [ ] Integrate notifications (desktop or terminal bell) for configurable events (pane command changes, keywords).

### Phase 4 — Extensibility & Packaging *(future)*
- [ ] Expose a gRPC / HTTP API for external tooling (CI dashboards, bots).
- [ ] Package binaries via GoReleaser and distribute via Homebrew and Linux packages.
- [ ] Support plugin hooks to run custom scripts on snapshot refresh.

## Risks & Mitigations

- **Performance**: Frequent `capture-pane` calls can be expensive with many panes.  
  *Mitigation*: cache content per pane with timestamps, allow user-tunable poll/capture intervals, and skip captures for hidden panes.

- **Permission / Environment**: tmux must be available in `$PATH` and compatible (3.1+).  
  *Mitigation*: detect missing binaries early, provide helpful error messaging, and document requirements prominently.

- **Viewport Drift**: Continuous capture-pane updates can fight with user scrolling.  
  *Mitigation*: keep cards auto-following new output only when the user is at the bottom; preserve manual scroll positions otherwise.

- **Snapshot Accuracy**: Restoring sessions requires faithfully reproducing commands, paths, and layouts.  
  *Mitigation*: when implementing snapshots, serialize layout strings, current commands, cwd, and optionally environment variables.

## Terminology

- **Session Sidebar**: Left-hand column listing `session:window` entries.
- **Pane Tabs**: Horizontal list above the content pane representing visible tmux panes for the selected window.
- **Hidden Pane**: Pane removed from the current view but still running; restored via `H`.
- **Capture Lines**: Number of lines to read from tmux history buffer when refreshing pane content.

## Implementation Notes

- Bubble Tea program runs in alt screen to take over the terminal and provide clean exit with `q` / `Ctrl+C`.
- Pane contents refreshed on each snapshot and selection change; future optimization might use tmux hooks or events.
- Reordering tabs is UI-only: we swap order in `paneOrder[windowID]` without modifying tmux layout.
- Keyboard map matches tmux/vim habits; forthcoming features should preserve mnemonic consistency.

## Success Metrics

- **Adoption**: number of developers running tmuxwatch alongside tmux daily.
- **Responsiveness**: pane content refresh under target (e.g., <= 500 ms with default interval).
- **Stability**: zero crashes across long-running sessions (>24 hours).
- **Feature Parity**: achieve core features of reference tools (e.g., tmux-tui) while maintaining Go simplicity.

## Open Questions

- Should hidden panes still capture output (for backlog) or only when unhidden?  
- Is there value in supporting remote tmux sessions over SSH via `wish` or `glow` integration?  
- What is the best way to persist user preferences (toml config, env vars, per-session state)?

Contributions should align with this roadmap, keeping code modular (`internal/tmux`, `internal/ui`, future `internal/store`) and maintaining clear documentation updates alongside features.
