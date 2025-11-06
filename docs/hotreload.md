# Hot Reload & Auto-Run Support

> Draft specification outlining the new “watch and rerun” workflows for Poltergeist and the `polter` CLI.

## Goals

- **Watch-mode for `polter`**: add a first-class `--watch` flag that keeps a target executable running and restarts it whenever Poltergeist completes a *successful* rebuild.
- **Config-driven auto-run**: allow `executable` targets to opt into automatic launch/relaunch directly from the Poltergeist daemon without needing the `polter` wrapper.
- **Event-friendly architecture**: surface build-success notifications in-process so both the new runner and third-party tooling can subscribe without polling state files.
- **tmux-friendly dev loop**: tmuxwatch will rely on this feature to keep its TUI session fresh while `poltergeist` handles background builds.

## Scope & Non-Goals

- ✅ Add CLI flags, configuration schema, runtime glue, docs, and regression tests.
- ✅ Ensure graceful shutdown/restart semantics (default `SIGINT`, configurable).
- ✅ Update `dist/` output (polter, poltergeist) once TypeScript changes land.
- ❌ Do not ship GUI integration yet (mac app changes postponed).
- ❌ No attempt to support multiple simultaneous auto-run processes per target; one runner per target is enough.

## High-Level Design

### 1. `polter --watch`

- Extend shared CLI options (`src/cli-shared/polter-command.ts`) with:
  - `-w, --watch` (boolean, default `false`)
  - `--restart-signal <sig>` (default `SIGINT`)
  - `--restart-delay <ms>` (default `250`)
- `ParsedPolterOptions` grows `watch`, `restartSignal`, `restartDelay`.
- `runWrapper` (`src/polter.ts`) branches:
  ```ts
  if (options.watch) {
    await runWithWatchMode({ target, projectRoot, options, args });
    return;
  }
  ```
- New helper `runWithWatchMode`:
  - Ensures daemon is running / build succeeded first (reuse existing checks).
  - Launch binary via extracted `launchTargetProcess(target, projectRoot, args, options)` (refactor from `executeTarget` so both call sites share code).
  - Watches the target’s state file using `fs.watchFile` (portable) and throttled JSON parsing:
    - On change, reload state.
    - If `lastBuild.status === "success"` and timestamp differs from last seen, restart child process (send configured signal, wait for exit, relaunch).
    - Ignore failures (show toast/log once, continue running the previous build).
  - Handles teardown on `SIGINT/SIGTERM` and when user presses `Ctrl+C`.
  - Optional enhancement: `--no-restart-on-failure` (default true).

### 2. Config-Driven Auto-Run

- Extend `ExecutableTarget` & schema (`src/types.ts`) with:
  ```ts
  autoRun?: {
    enabled?: boolean;          // default false
    command?: string;           // default: built binary
    args?: string[];
    restartSignal?: NodeJS.Signals;// default SIGINT
    restartDelayMs?: number;    // default 250
    env?: Record<string, string>;
  };
  ```
- Update `ExecutableTargetSchema` accordingly.
- `Poltergeist` target state adds optional `runner`.
  ```ts
  interface TargetState {
    ...
    runner?: ExecutableRunner;
  }
  ```
- Create `src/runners/executable-runner.ts`:
  - Accepts target metadata, manages a child process.
  - Methods: `start(initialBuildStatus)`, `stop()`, `onBuildStarted()`, `onBuildSucceeded(status)`, `onBuildFailed()`.
  - Maintains last-success timestamp to avoid duplicate relaunches.
  - Uses `spawn` with inherited stdio and optional env overrides.
- In `Poltergeist.start`:
  - If `target.autoRun?.enabled`, instantiate runner after builder validation (`state.runner = new ExecutableRunner(...)`).
  - After successful build inside `buildTarget`, `state.runner?.onBuildSucceeded(status)`.
  - On build failure, optionally surface message, keep previous process alive.
  - On daemon stop, `state.runner?.stop()`.

### 3. Event Hooks

- `StateManager` grows an `EventEmitter`.
  - `this.events.emit('status', targetName, buildStatus);` inside `updateBuildStatus`.
  - Expose `onStatus(listener)` and `offStatus`.
- `Poltergeist` runner can subscribe instead of reading files, but we still update the file for compatibility.
- Re-export emitter (or subscribe via `poltergeist.getStateManager()`).

### 4. Shared Launch Helpers

- Extract process launch logic (`executeTarget` / `executeStaleWithWarning`) into `launchBinary(options)`.
- Ensures both CLI watch mode and daemon runner use identical spawn semantics and error messages.

## Implementation Steps

1. **CLI plumbing**
   - Update `POLTER_OPTIONS` + parser + TypeScript types.
   - Refactor execute helpers; add `runWithWatchMode`.
   - Unit tests for option parsing & watch helper (using Vitest mocks).

2. **StateManager events**
   - Introduce emitter, add tests verifying emissions.

3. **ExecutableRunner**
   - New module with start/stop/restart logic & tests (mock child process via `child_process.spawn` stub).
   - Wire into Poltergeist core (start, buildTarget, stop).

4. **Schema & config**
   - Extend types & schemas (update `docs/API.md`, `docs/EXAMPLES.md`).
   - Provide default values & validation messages when misconfigured (e.g., autoRun without outputPath).

5. **Documentation**
   - `README` quick-start snippet for `polter target --watch`.
   - New section in docs about auto-run config & CLI flags.
   - Update CHANGELOG with feature summary.

6. **Distribution**
   - Run `pnpm build` (or `bun scripts/build-bun.js`) to refresh `dist`.
   - For tmuxwatch integration: add `poltergeist.config.json`, helper script, README notes.

7. **End-to-End test (manual)**
   - In tmux session:
     1. `pnpm build` (ensure poltergeist CLI uses new code).
     2. `poltergeist haunt --target tmuxwatch-cli` (new config).
     3. `polter tmuxwatch-cli --watch`.
     4. Edit Go file → confirm auto rebuild & tmuxwatch restart without manual intervention.

## Open Questions

- Do we auto-start the daemon if `polter --watch` detects it’s offline? (Initial plan: warn & exit; optional enhancement later).
- Should watch mode restart on *any* rebuild completion vs only success? (Current spec: success only; maybe add `--restart-on-failure` flag if needed).
- How do we surface build failures during watch mode? (Proposal: print summary + keep previous process alive; optionally tail logs on demand).
- For auto-run, do we allow multiple concurrently running commands per target? (Out of scope; only one runner per target).

---

This document will evolve as implementation progresses. Update the “Steps” checklist above as features land.
