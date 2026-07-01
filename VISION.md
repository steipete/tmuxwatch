# tmuxwatch Vision

tmuxwatch is a fast, dependable companion for understanding a local tmux estate at a glance. It should make active work, failures, and stale sessions obvious without forcing users to step through panes or surrender normal tmux workflows.

## Product principles

- **Monitor first.** Snapshot accuracy, low capture overhead, readable output, and long-running stability outrank feature breadth.
- **Keyboard first, mouse optional.** Every core navigation and focus workflow must remain efficient from the keyboard; mouse affordances may improve discovery but must not become required.
- **Observe by default.** Hiding, filtering, collapsing, and reordering are tmuxwatch view state. tmux sessions or processes change only after an explicit, clearly scoped user action; tmuxwatch never performs automatic cleanup.
- **Local and inspectable.** Prefer direct tmux commands, explicit state, and the existing `--dump` automation boundary. Add background services, remote access, network APIs, or plugin execution only for a demonstrated workflow that cannot stay local and simple.
- **Small configuration surface.** Prefer useful defaults and stable CLI flags. Add persistent configuration only after repeated user demand establishes a durable setting and its ownership.
- **Conservative persistence.** Saved layouts and history require a versioned, reviewable format, a preview before restore, and clear handling of commands, paths, environment data, and missing resources.
- **Lean implementation.** Keep the Go package boundaries clear, the Charmbracelet stack current, dependencies justified, and behavior covered close to its owning package.

## Priority order

1. Correct, efficient capture and resilient handling of missing, empty, or changing tmux state.
2. Fast navigation, search, focus, and readable status across many sessions and terminal sizes.
3. Safe, explicit session actions and better diagnostics for real tmux workflows.
4. User-controlled capture depth, refresh behavior, and theming when they preserve simplicity.
5. Workspace snapshot and restore only after the persistence safety contract is defined.

Remote tmux, general HTTP/gRPC APIs, notification systems, and arbitrary plugin hooks are later opportunities, not default roadmap commitments. They need a concrete user problem, a bounded security model, and proof that a smaller local integration is insufficient.

## Quality bar

Changes should preserve supported CLI behavior, include regression coverage when practical, pass formatting, lint, unit and race tests, and exercise the built binary against a real tmux session when runtime behavior changes. Packaging changes must keep Go, Nix, GoReleaser, and Homebrew paths reproducible.

Detailed architecture, terminology, and exploratory roadmap notes live in [`docs/spec.md`](docs/spec.md).
