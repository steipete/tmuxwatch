# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Added keyboard navigation for session cards (arrow keys + Enter) and a double-Esc chord to exit focus.
- Documented the Homebrew installation workflow in the README.
- Relaxed stale session detection to trust live preview activity and ignore attached sessions.

## [0.9] - 2025-11-05

- Initial implementation of the `tmuxwatch` CLI with Bubble Tea interface.
- tmux session/window/pane polling and auto-refresh.
- Added debug `--dump` mode for printing snapshots and hardened tmux timestamp parsing to accept missing data.
- Rebuilt the UI around session preview cards that capture active panes, auto-scroll, and share terminal space intelligently.
- Added live search/filter across sessions, windows, and panes.
- Mouse support: click to focus, scroll, and close cards; printable keys now forward to the focused tmux pane.
- Cards pulse briefly when new output arrives to highlight active panes.
- Headers now surface pane last activity timestamps and exit statuses.
- Ctrl+C now forwards to the live pane; press twice quickly to exit tmuxwatch (single press still quits if no live pane is focused).
- Added `.golangci.yml`, gofumpt formatting, `Makefile` helpers, and documentation for modern Go lint/format tooling.
- Integrated Bubble Tea viewport to clamp pane height, enable scrolling (ctrl+d/u, ctrl+f/b, g/G), and prevent oversized buffers from blowing up the layout.
- Documented tmuxwatch intent and primary use cases in the README.
