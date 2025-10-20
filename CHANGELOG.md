# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Initial implementation of the `tmuxwatch` CLI with Bubble Tea interface.
- tmux session/window/pane polling and auto-refresh.
- Key bindings for navigation, hiding panes, and manual refresh.
- Added debug `--dump` mode for printing snapshots and hardened tmux timestamp parsing to accept missing data.
- Pane tab view now captures live buffer contents via `tmux capture-pane`, supports keyboard tab swapping, and allows reordering with `[` and `]`.
