# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Initial implementation of the `tmuxwatch` CLI with Bubble Tea interface.
- tmux session/window/pane polling and auto-refresh.
- Added debug `--dump` mode for printing snapshots and hardened tmux timestamp parsing to accept missing data.
- Rebuilt the UI around session preview cards that capture active panes, auto-scroll, and share terminal space intelligently.
- Added live search/filter across sessions, windows, and panes.
- Mouse support: click to focus, scroll, and close cards; printable keys now forward to the focused tmux pane.
- Added `.golangci.yml`, gofumpt formatting, `Makefile` helpers, and documentation for modern Go lint/format tooling.
- Integrated Bubble Tea viewport to clamp pane height, enable scrolling (ctrl+d/u, ctrl+f/b, g/G), and prevent oversized buffers from blowing up the layout.
- Documented tmuxwatch intent and primary use cases in the README.
