**Idiomatic Go, 2025 edition — a practical guide**

**TL;DR**

* Prefer clarity over cleverness; lean on the standard library and the toolchain.
* Format with `gofmt`, lint with `staticcheck`, scan with `govulncheck`, and keep your module + toolchain current. ([Staticcheck][1])
* Use `context.Context` pervasively for cancellation/timeouts; log with `log/slog`; test with table-driven tests, fuzzing, and the race detector. ([go.dev][2])
* Learn the “new-ish” language/library bits you’ll meet in day‑to‑day code: `min`/`max`/`clear`, `slices`, `maps`, `cmp`, fixed `for`‑loop semantics (Go 1.22), and `math/rand/v2`. ([go.dev][3])

---

## 1) Style & philosophy

**What “idiomatic” means in Go**

* Small, composable packages; short names; avoid stutter (`bytes.Buffer`, not `bytes.ByteBuffer`); keep exported APIs minimal; comments start with the name they document. ([go.dev][4])
* Follow the Go team’s *Code Review Comments* for receivers, package names, contexts, error strings (“don’t capitalize, no trailing punctuation”), initialisms (`URL`, `ID`), and more. Bookmark this. ([go.dev][2])
* If you want an external style doc for team norms, Uber’s guide is widely used; treat it as advice, not law. ([GitHub][5])

**Do the basics automatically**

* `gofmt` (always), `go vet` (often), `staticcheck` (catches real bugs), and `govulncheck` (low‑noise vuln scanning). Wire all four into CI. ([Go Packages][6])

---

## 2) Modules, toolchains & workspaces (Go 1.18–1.25)

* **Modules**: Use a single `go.mod` per module. For v2+ modules, put the major version in the module path (`module example.com/thing/v2`). This is not a suggestion; it’s the rule. ([go.dev][7])
* **Toolchains** (Go 1.21+): The `go` command can auto‑use/download specific toolchains. The `toolchain` directive in `go.mod` pins builds for reproducibility; `GOTOOLCHAIN` lets you opt into auto‑fetch or keep it local. ([go.dev][8])
* **Workspaces** (`go work`): Great for multi‑module dev without sprinkling `replace` everywhere. Don’t ship your `go.work` to users; it’s for local dev flows. ([go.dev][9])

---

## 3) Language & stdlib features you should actually use

**Go 1.21 (Aug 2023)**

* Built‑ins: `min`, `max`, `clear`. ([go.dev][3])
* New stdlib: `slices`, `maps`, `cmp` — generic helpers that make everyday code cleaner (sorting, searching, equality, cloning). ([Go Packages][10])
* Structured logging: `log/slog` — use it for leveled, structured logs. ([Go Packages][11])

**Go 1.22 (Feb 2024)**

* **Fixed for‑loop gotcha**: Range/loop variables are fresh per iteration, so closures now “capture what you expect.” Less boilerplate copying inside loops. ([go.dev][12])
* `math/rand/v2`: Better API; prefer it for non‑crypto random. (For crypto use `crypto/rand`.) ([go.dev][13])

> Keep skimming the official release notes each cycle for incremental improvements (1.23, 1.24, 1.25). They’re concise and worth the five minutes. ([go.dev][14])

---

## 4) Errors & error design

**Ground rules**

* Return `error` as the last result; no exceptions for “expected” failures. Wrap inner errors with `%w` so callers can interrogate them. Prefer `errors.New` for static strings. ([Go Packages][15])
* Check with `errors.Is` / `errors.As` instead of string matching. Make sentinel errors (`var ErrX = errors.New("x")`) when consumers need to branch. ([Go Packages][16])
* Error messages: lower‑case, no trailing punctuation. Log context; don’t jam stack traces into error strings. ([go.dev][2])

**Example**

```go
var ErrNotFound = errors.New("widget not found")

func Load(id string) (*Widget, error) {
    w, err := repo.Get(id)
    if err != nil {
        return nil, fmt.Errorf("load %q: %w", id, err) // wraps
    }
    if w == nil {
        return nil, ErrNotFound
    }
    return w, nil
}
```

---

## 5) Contexts (timeouts, cancellation, request‑scoped values)

* Accept `context.Context` as the first parameter in methods that do I/O, block, or are request‑bound. Don’t store contexts in structs; don’t pass cancel funcs around; keep values small and typed. Use `WithTimeout/WithDeadline`. ([go.dev][2])
* Go 1.20 added `context.WithCancelCause` + `context.Cause` for better diagnosis in shutdown paths. Use it to propagate *why* you canceled. ([go.dev][17])
* Standard libs are context‑friendly: `http.Request.Context()`, `database/sql`’s `QueryContext`, etc. ([go.dev][18])

---

## 6) Concurrency: goroutines, channels, atomics

* Prefer simple fan‑out/fan‑in with `errgroup.Group` — cancels the group on first error, keeps code tidy. Use it with contexts for timeouts and shutdown. ([Go Packages][19])
* Reach for atomics **sparingly**; prefer `sync.Mutex` unless you’re sure. If you do use atomics, use the typed ones (`atomic.Int64`, `atomic.Pointer`) introduced in recent releases and remember the memory model basics. ([go.dev][20])

**errgroup sketch**

```go
g, ctx := errgroup.WithContext(ctx)
for _, u := range urls {
    u := u
    g.Go(func() error { return fetch(ctx, u) })
}
if err := g.Wait(); err != nil { /* handle */ }
```

([Go Packages][19])

---

## 7) HTTP services that behave well

* **Server timeouts**: set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`. These protect you from slowloris/abusive connections and accidental hangs. ([Go Packages][21])
* **Graceful shutdown**: call `srv.Shutdown(ctx)` on interrupt; tie it to context deadlines and cancellation causes. ([VictoriaMetrics][22])
* **Clients**: prefer a shared `http.Client` with sane `Timeout` and tuned `Transport`. (Avoid per‑request clients.)

**Minimal, production‑friendly server**

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Structured logs
    log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 2 * time.Second,
        ReadTimeout:       5 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    go func() {
        log.Info("listening", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Error("server error", "err", err)
        }
    }()

    <-ctx.Done()
    shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutCtx) // graceful
}
```

([Go Packages][11])

> Instrumentation: wrap handlers with OpenTelemetry’s `otelhttp` for traces/metrics without changing your business logic. ([Go Packages][23])

---

## 8) Logging & observability

* Prefer `log/slog` for structured logs; add fields, groups, and levels; choose JSON for machines, text for dev. You can write or wrap handlers for advanced use cases (sampling, redaction). ([Go Packages][11])
* For tracing/metrics, OpenTelemetry’s Go SDK is the de‑facto standard and integrates cleanly with `net/http` via `otelhttp`. ([OpenTelemetry][24])
* Profile when performance matters: `runtime/pprof` and `net/http/pprof` are one import away. Use the tooling (`go tool pprof`) and keep `/debug/pprof` protected. ([Go Packages][25])

---

## 9) Testing that scales with the codebase

* **Table‑driven tests** + subtests keep coverage dense and readable. Put fixtures under `testdata/`. ([go.dev][26])
* **Fuzzing** (Go 1.18+) is built into `testing`. Use it on parsers, decoders, and anything with edge‑casey inputs. ([Go Packages][27])
* Run **the race detector** (`go test -race`) on CI for packages that touch concurrency. (It’s worth the CPU cycles.)
* Golden tests are fine for big outputs; gate updates (`-update`) and always review diffs. ([Go Packages][28])

**Tiny table‑driven test**

```go
func TestSlug(t *testing.T) {
    cases := []struct{
        in, want string
    }{
        {"Hello, World!", "hello-world"},
        {" café  ", "cafe"},
    }
    for _, tc := range cases {
        t.Run(tc.in, func(t *testing.T) {
            if got := Slug(tc.in); got != tc.want {
                t.Fatalf("got %q want %q", got, tc.want)
            }
        })
    }
}
```

([go.dev][26])

---

## 10) Practical language tips (2025)

* **Loops**: In Go 1.22+, closures capture the per‑iteration variable; your old `v := v` workaround is no longer needed in code compiled as 1.22+. Still be mindful when you *take addresses* of loop vars. ([go.dev][12])
* **Generics**: Keep constraints simple; prefer concrete types at package boundaries (“accept interfaces, return structs”). Use `cmp.Ordered` instead of third‑party “constraints”. Lean on `slices` helpers over custom loops. ([go.dev][3])
* **Randomness**: non‑crypto → `math/rand/v2`; crypto → `crypto/rand`. ([go.dev][13])
* **Logging**: standardize on `slog` and wire your handler of choice; don’t reinvent structured logging. ([Go Packages][11])

---

## 11) Security & supply chain

* Use `govulncheck` locally and in CI to flag *reachable* vulnerabilities with call stacks; far less noisy than CVE scans that don’t know your code path. ([go.dev][29])
* Keep the `go` and `toolchain` lines in `go.mod` current so teammates and CI use compatible compilers and stdlibs. ([go.dev][8])

---

## 12) A tiny, idiomatic project layout

```
.
├── cmd/
│   └── api/          # main package for the service binary
├── internal/         # private packages (not importable by others)
│   ├── httpx/        # handlers, routers
│   ├── store/        # database access
│   └── domain/       # business types & logic
├── pkg/              # optional: only for APIs meant to be imported by others
├── go.mod
└── go.work           # local workspace (do not publish)
```

* `cmd/<app>` holds `main`. `internal/` is for everything you don’t want imported by other modules. Keep `pkg/` only for public APIs you’re committed to. ([go.dev][4])

---

## 13) Checklists you can paste into a PR template

**HTTP service**

* [ ] Context on all handlers and outbound calls
* [ ] Server timeouts set; graceful shutdown implemented
* [ ] `slog` structured logs with request IDs
* [ ] `otelhttp` wrapping + metrics/traces exported (if you use OTel)
* [ ] `pprof` only enabled behind admin auth or non‑prod builds
* [ ] `go vet`, `staticcheck`, `govulncheck` clean ([Go Packages][21])

**Library**

* [ ] Package doc comment present; no exported stutter
* [ ] Small interfaces; return concrete types
* [ ] Errors wrap inner failures; `errors.Is/As` in tests
* [ ] Generics only where they clearly help (prefer `slices`/`maps` helpers) ([go.dev][4])

---

## 14) Snippets you’ll reuse

**slog setup**

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
logger.Info("started", "commit", commitSHA)
```

([Go Packages][11])

**Context with cause**

```go
ctx, cancel := context.WithCancelCause(parent)
deper cancel()
// ...
cancel(fmt.Errorf("shutting down: %w", ErrSIGTERM))
if err := context.Cause(ctx); err != nil { /* explain why */ }
```

([go.dev][17])

**Using `slices` and `cmp`**

```go
s := []int{3,1,4}
slices.SortFunc(s, cmp.Less[int])
top := max(s...) // built-in
clear(s)         // built-in
```

([Go Packages][10])

---

### Keep learning (authoritative sources)

* *Effective Go* and *Code Review Comments* (style and idioms). ([go.dev][4])
* Release notes for 1.21–1.25 (language + stdlib changes). ([go.dev][3])
* Tooling guides: toolchains & workspaces; staticcheck; govulncheck. ([go.dev][8])
* Observability: OpenTelemetry for Go; `otelhttp`. ([OpenTelemetry][24])
* HTTP timeouts & shutdown: `net/http` docs + deep dives. ([Go Packages][21])

---
