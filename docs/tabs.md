# Tabs Roadmap

This note captures the work required to integrate BubbleApp’s Tab controls into tmuxwatch.

## Goals
- Top-level tab bar driven by `component/tabs` so users can switch between the grid summary and a single-session detail view.
- Per-session controls for **maximize** (switch the Tab to detail mode) and **collapse** (shrink the card in the grid).
- Keyboard ergonomics: `shift+left/right` to change tabs, `enter` to pop into detail, `esc` to return.

## BubbleApp Integration
1. Vendor the `github.com/alexanderbh/bubbleapp` module and initialise a lightweight `app.Ctx` that renders only the tab strip. BubbleApp already exports `tabs.New` and `tabtitles.New` widgets that manage focus, hover, and theming.
2. Mount the tab component inside the existing Bubble Tea model by:
   - Maintaining the active tab index in `ui.Model` (`viewMode` enum).
   - Delegating tab key/mouse events to BubbleApp’s handlers via a small adapter (`tabsBridge.Update(tea.Msg)`).
   - Rendering the strip with `tabsBridge.View()` and injecting it at the top of `View()` before the title bar.
3. Theme alignment: Map tmuxwatch palette to BubbleApp’s theme struct so the tab colours follow our existing Lip Gloss scheme.

## Layout Changes
- **Grid Tab**: Preserve current card layout, but add per-card action glyphs (`[□]` maximize, `[–]` collapse). Collapsed cards render the header only, freeing vertical space.
- **Detail Tab**: Reuse `sessionPreview` viewport at terminal width/height; surface pane variables and stale state in a sidebar. Provide `c`/`[` shortcuts to collapse/return.
- Update `updatePreviewDimensions` to honour collapsed heights and full-screen detail mode.

## Risks & Mitigations
- BubbleApp expects to own the Bubble Tea program; we’ll isolate it inside an adapter so we only use the tab widget and keep existing state management.
- Mouse zone IDs must be synchronised between tmuxwatch and BubbleApp to avoid conflicting hit-tests.
- Performance: limit pane capture to visible sessions (already handled) and gate detail polling to the active tab.

## Next Steps
1. Spike a `tabsBridge` wrapper that initialises BubbleApp, handles `tea.Msg`, and returns the rendered tab strip.
2. Add model fields for `tabsActive`, `viewMode`, and collapsed state tracking.
3. Wire maximize/minimize actions and update keyboard/mouse handlers.
4. Expand tests to cover collapsed layout calculations and tab switching logic.
