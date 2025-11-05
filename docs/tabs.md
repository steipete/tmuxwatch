# Tab Bar Integration

This note records how the new tabbed layout is wired into tmuxwatch.

## BubbleApp component usage
- We render the strip with `github.com/alexanderbh/bubbleapp/component/tabtitles`. The BubbleApp runtime handles the layout and styling of the tab titles, while tmuxwatch keeps ownership of state and input routing.
- The component is instantiated on each `View` render through a lightweight helper that calls `app.NewCtx()` and renders `tabtitles.New(...)` for the current tab titles. We treat the BubbleApp output purely as view markup; key and mouse handling stays in tmuxwatch so the existing Bubble Tea model remains unchanged.
- BubbleZone IDs emitted by the component are inspected in `handleTabMouse`; clicks on `tab:<n>` now switch tabs (with helper coverage in `handlers_test.go`).

## Tabs & view modes
- **Overview tab** (always present) renders the existing grid layout.
- **Session tab** appears when a session is maximised. We use a new `viewModeDetail` flag to render only the selected session, reusing the existing card renderer.
- `shift+left/right` switches tabs. The maximize control (`ctrl+m` or clicking `[^]`) jumps straight to the session tab; `esc` or the Overview tab returns to the grid.

## Card controls
- Headers now expose three affordances: `[ ^ ]` maximise/restore, `[-]/[+]` collapse, `[x]` hide. We map those to mouse zones via BubbleZone so clicks continue to work alongside the new keyboard shortcuts (`ctrl+m`, `z`, `Z`).
- Collapsed cards collapse their body content to a single header row so the grid can show more sessions at once.

## Implementation checklist
- Track additional UI state in the Bubble Tea model: `viewMode`, `detailSession`, `activeTab`, and a `collapsed` set.
- Adjust `filteredSessions` to return only the detail session when in detail mode.
- Update the status footer and cheat sheet to describe the new controls.
- Cover the helper logic with unit tests (`state_test.go`).

With these pieces in place, the tabs integrate cleanly while tmuxwatch retains its single Bubble Tea program and existing bubblezone layouts.
