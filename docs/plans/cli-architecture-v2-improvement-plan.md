# CLI Architecture v2 Improvement Plan — cobra/viper, Interactive TUI, OpenTelemetry

Status: **PLAN ONLY — no implementation yet.** Created 2026-07-16 from a three-track deep
analysis of `cmd/`, `internal/cli/`, `internal/flow/`, `internal/schedule/`, `internal/app/`,
`internal/connectors/`, `internal/worker/`, plus web research on the spf13 (cobra/viper),
charmbracelet (Bubble Tea v2 ecosystem), and OpenTelemetry Go states of the art (July 2026).
This document records findings and the proposed program; it does not change any runtime file.
Companion documents:

- `docs/design/tui-ux-design.md` — full UX/UI design for the interactive layer (Pillar B).
- `docs/adr/0002-cobra-viper-cli-framework.md` — framework decision record.
- `docs/adr/0003-interactive-tui-layer.md` — events bus + TUI decision record.
- `docs/adr/0004-opentelemetry-observability.md` — observability decision record.
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` — pi/GSD execution prompt.

---

## 1. Research basis

Sources reviewed:

1. **spf13/cobra** v1.10.x + pflag semantics (`github.com/spf13/cobra`, cobra.dev) —
   `DisableFlagParsing`, `SilenceErrors/SilenceUsage`, `SetHelpFunc`/`SetUsageFunc`,
   `SetOut/SetErr/SetArgs`, `ValidArgsFunction`, `FParseErrWhitelist`, `cobra/doc.GenManTree`.
2. **spf13/viper** — instance-scoped `viper.New()`, `SetEnvPrefix`/`BindEnv`/`BindPFlag`,
   `SetConfigFile`, `Unmarshal`; transitive dependency set (afero, cast, fsnotify, gotenv,
   pelletier/go-toml, mapstructure, sourcegraph/conc).
3. **Charm v2 line** (stable 2026-02-23, `charm.land` import paths; verified July 2026):
   bubbletea v2.0.8 (declarative `tea.View`, `KeyPressMsg`, `tea.RequestBackgroundColor`),
   bubbles v2.1.1 (list/table/spinner/progress/viewport/help/key), lipgloss v2.0.5
   (`LightDark`, colorprofile degradation), **huh v2.0.3** (form groups, dynamic
   `OptionsFunc`/`WithHideFunc`, **`WithAccessible` screen-reader mode**), **glamour v2.0.1**
   (terminal markdown), colorprofile v0.4.3 (NO_COLOR/CLICOLOR/TERM=dumb),
   Evertras/bubble-table v0.22.3 (data grid; already on charm.land v2), teatest/v2.
4. **gh CLI accessibility work** (github.blog "Building a more accessible GitHub CLI";
   accessibility.github.com CLI guide) — accessible prompter built on huh,
   `accessible_colors`, spinner-disable, `gh a11y` topic.
5. **clig.dev** — prompts-as-progressive-enhancement doctrine, `--no-input`, TTY-aware output.
6. **OpenTelemetry Go** v1.44.0 / v0.66.0 / v0.20.0 (May 2026 train) — traces + metrics
   stable, logs beta; stdout exporters with `WithWriter`; `otelslog` bridge v0.19.0;
   `go.temporal.io/sdk/contrib/opentelemetry` v0.7.0; OTLP http/protobuf vs grpc weights.
7. This repo: `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md` (§9 config, §15
   observability, §16 errors, §17 dependencies), `AGENTS.md` (agentic contract, parity gate,
   go.mod human gate), `CONTEXT.md`, `internal/cli/agentic_contract_test.go`.

### Distilled best practices (the yardstick)

**A. Flags are the API; prompts are progressive enhancement.** (gh, clig.dev) Prompt only
when stdout+stdin are TTYs and a required input is missing; `--no-input` disables all
interactivity and errors must name the exact flag to pass; every wizard result is
expressible as flags and the wizard teaches them.

**B. A CLI framework should own routing, help, and completion — not behavior.** (cobra
practice) Behavior stays in testable handlers with injected writers; the framework layer is
thin and replaceable.

**C. Config wants one typed struct with explicit precedence** (flags > env > file >
defaults), resolved once per invocation, injected — never ambient globals or file watching
in a deterministic CLI. (viper used in instance mode)

**D. TUIs must be gated, degradable, and honest.** TTY-only activation; colorprofile
degradation TrueColor→256→16→none honoring NO_COLOR/CLICOLOR/TERM=dumb; color never the
only carrier of meaning; accessible (sequential, no-redraw) mode for screen readers;
final frames tell the truth about cancellation/failure. (charm, gh a11y)

**E. Progress is an event stream, not a poll loop.** Long-running operations emit typed
events through a dependency-free bus; UI, agents (NDJSON), and files are just sinks.

**F. Observability layers are siblings, not derivatives.** Real-time events (UI), durable
ledger (audit), and traces/logs/metrics (diagnostics) instrument the same call sites and
correlate by run ID, but never depend on each other.

**G. Telemetry in a CLI is opt-in, secret-safe, and exit-code-neutral.** Default = no SDK
constructed (zero flush latency); attribute allowlists + value-registry redaction; never
record bodies/headers/query strings; init/flush failures warn, never fail the command.

**H. Never regress the machine contract while improving the human one.** Golden transcripts
recorded before migration; JSON envelopes, exit codes, and help text stay byte-identical
until a deliberate, parity-gated change.

---

## 2. Current-state assessment (what is already good)

- **The agentic contract is real and enforced.** `--json` API-versioned envelopes
  (`internal/cli/parse.go:128-139`), stderr-for-diagnostics, no-ANSI via
  `safety.SanitizeTerminal` (`internal/safety/safety.go:32`), all asserted by
  `internal/cli/agentic_contract_test.go`. (Matches H — there is a contract worth guarding.)
- **Error taxonomy → stable exit codes** (usage=2, validation=3, auth=4, connector=5,
  runtime=6, policy=7, internal=1; `internal/cli/errors.go:107-127`) with a single
  `writeError` funnel — exactly the shape a framework migration wants to preserve.
- **Everything is `io.Writer`-injected** (`cli.Run(args, stdout, stderr)`), so the whole CLI
  is testable in-process (7k+ lines of CLI tests) and the certify harness can drive it
  recursively (`certify.SetCLIRunFunc`, `cmd/pm/main.go:22`). (Matches B.)
- **Domain enums a wizard needs already exist as code**: sync modes with
  `RequiresCursor()`/`RequiresPrimaryKey()` (`internal/app/sync_modes.go`), flow step kinds
  (`internal/flow/manifest.go:12-17`), schedule backends (`internal/schedule/schedule.go:28-33`),
  connector registry with capability metadata (`internal/connectors/bundleregistry`),
  a currently-unused cron `Next()` (`internal/schedule/cron.go:132`).
- **Connector manuals render from structured data on demand**
  (`connectors.RenderConnectorManual`, `internal/connectors/guide.go:224`) — a docs-viewer TUI
  has a clean text source with zero new plumbing.
- **The single HTTP chokepoint exists** for connector traffic (`connsdk.Requester.do`,
  `internal/connectors/connsdk/http.go:216`) — one place to instrument. Flow steps already
  record `DurationNs`; per-run log directory `.polymetrics/logs/` is already created
  (`internal/app/app.go:57-71`).
- **`make verify` is a genuine gate** (fmt, tidy-check, vet, 20m tests, build, docs-check,
  smoke, lint, connectorgen-validate) — every phase below ends on it.

---

## 3. Gap analysis

Each gap cites the offending file(s) and the yardstick letter it violates.

### 3.1 No CLI framework — routing, help, and completion are all hand-rolled — B, H

- Dispatch is a 21-case `switch` (`internal/cli/cli.go:57-100`) with nested per-namespace
  switches (~60+ paths). Flag parsing is bespoke and untyped
  (`parseFlags` → `map[string][]string`, `internal/cli/parse.go:47-67`): `--flag` with no
  value means `"true"`, unknown flags are silently collected, `--flag v` binds only if the
  next token doesn't start with `--`.
- Help is hand-authored strings (`internal/cli/docs.go:3-90`); `docs/cli/**` is generated
  from the same map; the website reference is maintained separately. No shell completion,
  no `pm etl run --help` focused help (you get the whole namespace page), no man pages.
- The PRD §17 explicitly calls for a CLI framework; it was never adopted.
- **Effect:** every new command hand-implements parsing/help/errors; agents and humans get
  no completion; help drift between the three surfaces is manual work.

### 3.2 Configuration is scattered and the config file is write-only — C

- `.polymetrics/config.yaml` is written by `pm init` (`internal/app/app.go:76-82`) and read
  by nothing. ~11 env vars are read ad hoc via `os.Getenv` across six packages
  (`internal/runtimecheck/runtimecheck.go:36-43`, `internal/cli/worker_cli.go:30`,
  `internal/schedule/select.go:22`, `internal/schedule/crontab.go:23,64`,
  `internal/cli/agent_image_cli.go:17,24`, `internal/worker/submit.go:82-86`,
  `internal/cli/rlm_cli.go:110-126`). Two prefixes coexist (`POLYMETRICS_*`, `PM_*`).
- **Effect:** no single place to see or document configuration; per-project settings can't
  be persisted; precedence is undefined.

### 3.3 No event stream — progress is polled from files — E

- ETL returns only final counts (`internal/app/app.go:460-560`); flow engine returns
  `RunResult` at the end and progress is observed by reading a checkpoint file
  (`internal/flow/checkpoint.go`) via a separate `pm flow status`; certify batch persists
  `progress.json` at completion (`internal/connectors/certify/batch.go:273,322-341`);
  the RLM Temporal submit blocks in `run.Get` (`internal/worker/submit.go:53-60`) and
  heartbeats never reach the user.
- **Effect:** neither humans nor agents can watch a run; a TUI has nothing to subscribe to.

### 3.4 The human UX layer does not exist — A, D

- **No `pm flow create`.** Flows are hand-authored JSON (`internal/flow/manifest.go:31-51`)
  where dependencies are *implicit* — inferred by matching `in`/`out` table names
  (`internal/flow/dag.go`). One typo in a table name silently rewires or invalidates the DAG.
- **`pm schedule create`** demands a hand-written 5-field cron string
  (`internal/schedule/cron.go:21`); `--flow` is stored unvalidated; `install` is a separate
  step; `schedule.Next()` exists but is dead code (no "next fire times" preview).
- **`pm query run`** takes SQL only via `--sql`; results are raw NDJSON
  (`internal/cli/cli.go:930-933`); **no command enumerates queryable warehouse tables**
  (the DuckDB engine derives them from `warehouse/*.jsonl`,
  `internal/app/query_engine_duckdb.go:131`, but users must guess).
- **`pm connectors list`** prints **551 rows** with a raw `%+v` struct dump
  (`internal/cli/cli.go:186-188`) — the largest browse surface in the product, unusable.
- **Docs are plain text only** — no markdown rendering, no pager; `pm docs` only
  generates/validates files on disk (`internal/cli/cli.go:1118`).
- **Reverse ETL** requires copy-pasting a plan ID and an approval token across four
  commands; the JSON output deliberately redacts the token
  (`safeReversePlanForOutput`, `internal/cli/cli.go:858`) so it can only be lifted from
  human text output.
- **`pm flow status` defaults its flows dir to `os.TempDir()`**
  (`internal/cli/flow_cli.go:204`) while `run` uses `<projectDir>/flows` and `list` uses
  `.polymetrics/flows` — a real inconsistency.
- **Effect:** the product is agent-first by design but human-hostile by accident; every
  creation journey is hand-authoring files and relaying IDs between commands.

### 3.5 Zero observability — F, G

- No logging framework anywhere (no slog/zerolog/stdlib log in non-test code); Temporal
  gets a `noopLogger` (`internal/worker/submit.go:99-106`). No metrics, no tracing, no OTel.
  PRD §15 (per-run JSONL logs, metrics list, redaction middleware) is unimplemented.
- **Effect:** a failed run leaves nothing behind except a one-line ledger row; performance
  questions (why is this connector slow, how many retries) are unanswerable.

---

## 4. Proposed improvements (three pillars, phased)

### Pillar A — cobra + viper via strangler migration (fixes 3.1, 3.2)

**Strategy: cobra becomes the router on day one; every existing handler and the bespoke
flag parser keep running unchanged underneath** (`DisableFlagParsing: true` wrappers), then
namespaces are promoted to native cobra parsing one GSD loop at a time. Rejected
alternatives (big-bang rewrite, cobra-as-fallback, fang) are recorded in ADR-0002.

Key mechanics (full detail in ADR-0002):

- Root command: `Use: "pm"`, `Args: cobra.ArbitraryArgs`, `DisableFlagParsing: true`,
  `SilenceErrors/SilenceUsage: true`; persistent `--root`/`--json` definitions; root `RunE`
  fallthrough preserves the dynamic `pm <connector> <path…>` dispatch and the exact legacy
  `unknown command %q` error. A **fresh command tree per `Run()` call** keeps the certify
  in-process harness and parallel tests safe. `cli.Run`'s signature does not change.
- `mapCobraErr` classifies cobra/pflag errors into the existing taxonomy so `writeError`
  stays the only place computing exit codes.
- Custom `SetHelpFunc`/`SetUsageFunc` render the existing `docs` map (man-page text), honor
  the `--json` `CommandManual` envelope, and keep bare-namespace-help-exit-0. The docs map
  remains the single help source until the deliberate help-tree phase.
- Promotion recipe per namespace: declared pflags with `StringArrayVar` (repeated-flag
  accumulation, not comma-split), `NoOptDefVal="true"` where bare `--flag` means true,
  `FParseErrWhitelist{UnknownFlags: true}` to keep legacy tolerance until a documented
  tightening. Certify subtree migrates last; connector dispatch never migrates.
- **Viper in instance mode** (user decision; discipline rules in ADR-0002): `viper.New()`
  inside `config.Load` — never the package-level singleton — `SetEnvPrefix("POLYMETRICS")`,
  explicit `BindEnv` per key with `PM_*` legacy aliases, `BindPFlag` for global flags,
  `SetConfigFile(.polymetrics/config.yaml)` (giving the file its first reader), **no
  `WatchConfig`**, `Unmarshal` into a typed `Config` struct injected down the call tree.
  Precedence: flags > env > file > defaults. `credentials add --from-env` and certify's
  credsfile keep raw `os.Getenv` — they read *user-named* env vars (data, not config).
- UX wins: shell completion (`ValidArgsFunction` for connector/credential/connection names,
  returning `ShellCompDirectiveNoFileComp` silently when no project exists), help tree,
  `cobra/doc.GenManTree` via a new `pm docs man`, typo suggestions appended to the existing
  unknown-command message.
- **Golden transcript safety net first**: ~80 recorded invocations (exit code + stdout +
  stderr) covering every help path, both flag forms, repeated flags, unknown-flag
  tolerance, JSON envelopes, error categories, connector-dispatch errors — recorded against
  the legacy dispatcher and kept green through every phase. Intentional changes (help tree)
  are reviewed fixture diffs.

### Pillar B — Interactive UX layer: events bus + Bubble Tea v2 (fixes 3.3, 3.4)

Full design with wireframes, palette, keybindings, and accessibility spec in
`docs/design/tui-ux-design.md`; decision record in ADR-0003. Summary:

- **`internal/events` first** (stdlib + `internal/safety` only): typed `Event`
  (kind/scope/run/step/status/counters), `Emitter` via context
  (`events.WithEmitter`/`FromContext`, `Nop` default), sinks: `NDJSON` (→ stderr; live
  progress **for agents** via a new `--progress ndjson` flag — value before any TUI
  exists), `Chan` (TUI bridge; lifecycle events never dropped, progress coalescible),
  `Throttle`, `Multi`. Instrumentation points: `flow.Engine.Run` beside the existing
  `appendLedger` calls, ETL per-batch `flush()`, certify worker pool, and a Temporal
  `DescribeWorkflowExecution` poller in `worker/submit.go` (no workflow code changes).
- **TTY gate** `ui.Detect`: TUI only when stdout is a TTY ∧ ¬`--json` ∧ ¬`--plain` ∧
  ¬`--no-input` ∧ `PM_NO_TUI`/`CI` unset ∧ `TERM≠dumb`. New `cli.RunWithOptions`; existing
  `Run` delegates with `ModePlain`, so every existing test exercises the plain path by
  construction. `SanitizeTerminal` continues to guard the plain path; on the TUI path it
  moves into view-string hygiene (every dynamic string sanitized before styling).
- **Doctrine (A):** flags stay the API. Wizards prompt only for missing inputs, print the
  equivalent non-interactive command on completion, and `--no-input` errors name the flag.
- **Surfaces** (each with plain/JSON parity): live run dashboards (`flow run`, `etl run` —
  the signature "pipeline rail"), `pm flow create` wizard (huh dynamic forms; `in`/`out`
  wiring becomes structurally correct because pickers only offer tables produced upstream),
  `pm schedule create` interactive (cron presets + next-3-fire-times preview using the
  dead `schedule.Next`), `pm query tables` (plain enumerator — agents and wizard share it)
  + interactive query grid (bubble-table), connectors browser (551-row fuzzy list + manual
  preview pane), docs viewer (glamour pager over existing manual text), certify batch
  table, RLM viewer, guided reverse-ETL session (token relay handled in-session).
- **Accessibility from day one** (D): huh `WithAccessible` wired to `--accessible` +
  `PM_ACCESSIBLE_PROMPTER`/`ACCESSIBLE`; spinner-disable; 4-bit accessible palette option;
  color always paired with glyph+word; min-size guard; `pm a11y` help topic.

### Pillar C — Observability: slog first, then OpenTelemetry (fixes 3.5)

Decision record in ADR-0004. Summary:

- **Phase order: logs → traces → metrics → (optional) otel log bridge.** Logging is
  stdlib-only (`log/slog`), default ON, and delivers PRD §15.1 immediately: a
  `RedactingHandler` (key-based redaction from connector `SecretFields` + a value registry
  fed from the `vault.Get` chokepoint, `internal/vault/vault.go:80`, + existing
  `safety.RedactErrorText`) fans out to per-run `.polymetrics/logs/<run-id>.jsonl`
  (ctx-routed) and stderr at warn+. Temporal's `noopLogger` is replaced with
  `tlog.NewStructuredLogger` (ships with the pinned SDK — zero new deps).
- **Traces** (default OFF; exporters `none`/`file`/`otlp`): `internal/telemetry`
  Init/Shutdown (3s bound; disabled path constructs no SDK), spans `pm.command` →
  `pm.etl.run`/`pm.flow.step`/`pm.certify.connector` → `pm.connector.http` instrumented
  **directly in `connsdk.Requester.do`** — otelhttp is rejected because it records
  `url.full` and `api_key_query` auth puts credentials in query strings; we record
  scheme+host+path only, one span per logical request with per-attempt events.
- **Metrics** per PRD §15.2 with a hard hot-loop rule: local counters flushed per batch
  flush, never per-record instrument calls (benchmark guard). Temporal tracing interceptor
  + metrics handler from `contrib/opentelemetry`, gated on enabled.
- **Layering (F):** events bus / run ledger / OTel+slog are siblings at the same call
  sites, correlated by `run_id`; none derives from another.
- **Safety (G):** attribute-key allowlist test; value-registry scrub; never bodies,
  headers, argv, or query strings; red test = smoke flow with a known token, grep logs +
  telemetry for absence (with a test hook proving the grep can fail).

---

## 5. Consolidated roadmap (one GSD loop per phase)

Tracks interleave after phase 2; A3/A4 unblock C-track config keys; B-track needs phase 7
before any TUI phase.

| # | Phase | Track | New go.mod deps (human gate) |
|---|-------|-------|------------------------------|
| 1 | Golden transcript safety net + docs-generate-diff test | A | — |
| 2 | Cobra router shell (strangler, byte-identical) | A | cobra (+pflag, mousetrap) |
| 3 | `internal/config` with viper + config.yaml reader | A | viper (+transitives) |
| 4 | Migrate scattered env reads onto config | A | — |
| 5 | `internal/events` bus + engine/ETL/certify/worker instrumentation | B | — |
| 6 | slog foundation + redaction + per-run JSONL logs + Temporal logger | C | — |
| 7 | TTY gate, `--plain`/`--no-input`/`--progress ndjson`, `internal/ui/styles` foundation | B | golang.org/x/term |
| 8 | Nativize pilot namespace (`catalog`) | A | — |
| 9 | Nativize remaining namespaces (several loops; certify last) | A | — |
| 10 | Flagship run dashboards: `flow run` + `etl run` | B | bubbletea/bubbles/lipgloss v2, teatest (test-only) |
| 11 | Wizards wave 1: `pm flow create` + `pm schedule create` (+ accessible mode) | B | huh v2 |
| 12 | Traces: command→operation→connector HTTP + file/OTLP exporters | C | otel api/sdk/trace exporters |
| 13 | Browse wave: connectors browser, `pm query tables`, interactive query grid | B | evertras/bubble-table |
| 14 | Docs viewer: glamour pager for command + connector manuals | B | glamour v2 |
| 15 | Connector command registration + shell completion | A | — |
| 16 | Certify batch table + RLM agent viewer dashboards | B | — |
| 17 | Metrics per PRD §15.2 + Temporal otel contrib | C | sdk/metric, exporters, temporal contrib |
| 18 | Wizards wave 2: guided reverse-ETL session + credentials/connections prompting | B | — |
| 19 | Help tree deepening + man pages (the deliberate help-churn phase) | A | — |
| 20 | `pm a11y` topic + accessibility audit pass | B | — |
| 21 | OTel log bridge (beta; optional — droppable) | C | otel log (beta, pinned) |
| 22 | Cleanup: dead parse code, AGENTS.md/CONTEXT.md updates | A | — |

Every phase gates on `make verify` plus phase-specific checks (golden suite, `-race`,
secret greps, `git diff go.mod` expectations) — encoded per stage in the execution prompt.

---

## 6. Risk register

| Risk | Mitigation |
|------|------------|
| Help-text churn breaks the docs/website parity gate | Docs map stays the single help source through phase 18; custom help/usage funcs everywhere so cobra boilerplate can't leak; all deliberate churn quarantined in phase 19 with regenerated fixtures + website budget |
| pflag semantic drift (`--flag` w/o value, unknown-flag tolerance, comma-split, error wording, usage-vs-validation category) | Golden edge-case fixtures recorded in phase 1 before opinions form; `StringArrayVar` + `NoOptDefVal` conventions; `FParseErrWhitelist`; `mapCobraErr` pins categories |
| Certify in-process re-entrancy | Fresh command tree per `Run`; signature frozen; certify subtree on `DisableFlagParsing` until its own loop; `certify_exit_test.go` + smoke in every A-phase gate |
| Agent-contract regression from the TUI | Plain path is the compile-time default of untouched `cli.Run`; gate can never be satisfied by pipes/CI; NDJSON confined to stderr; import-direction CI check (`internal/ui` never imported by business packages); one contract test per TUI-enabled command |
| Dependency weight vs "dependency-free" tenet | Phases 5–6 (events, slog) ship value with zero new deps; heavy gates isolated to phases 10/11/12/13/14; default runtime behavior unchanged (telemetry off, TUI TTY-only); tenet interpretation recorded in ADR-0004 |
| Hot-loop overhead (per-record paths) | No per-record spans/logs/instrument calls anywhere; counters accumulate locally and flush per batch; benchmark guard in phase 17 |
| Bubble Tea v2 / otel-logs beta churn | v2 core is stable (powers charm's own apps); teatest pinned pseudo-version, test-only; otel log bridge confined to one file, pinned, explicitly droppable (phase 21) |
| Windows terminals | Inline mode (no alt screen) for run commands; ASCII glyph fallbacks; colorprofile degradation; `PM_NO_TUI` universal escape hatch |
| Event-bus backpressure stalls engine | Buffered `Chan` with bounded-timeout fallback + `Dropped()` accounting; runner goroutine always closes bus + sends `DoneMsg` (deferred) |
| Flush latency added to command exit | Disabled telemetry constructs no SDK and skips shutdown; enabled path bounded at 3s warn-and-continue |

---

## 7. Verdict

The three pillars are independent enough to interleave but compound: cobra gives every new
interactive command a natural home (`pm flow create`, `pm docs view`, `pm query tables`);
the events bus makes both the TUI and agent NDJSON progress possible from one
instrumentation pass; slog + OTel reuse the same call sites and the same redaction
machinery the TUI needs for view hygiene. The migration is deliberately boring where it
must be (byte-identical strangler phases guarded by golden transcripts) and bold in one
place (the interactive layer, where the product currently has nothing).

## 8. Non-goals / out of scope

- No interactive secret entry: `credentials add` keeps env/stdin intake only
  (`--from-env`, `--value-stdin`) — an interactive secret prompt would create a second,
  less-auditable intake path (recorded in `docs/design/tui-ux-design.md`).
- No REPL/shell mode; no daemon UI; no web UI.
- No prometheus exporter; no otelhttp; no grpc-exporter promotion (http/protobuf default).
- No renaming of existing flags/env vars — `PM_*` names remain honored as aliases.
- No changes to the reverse-ETL approval token model — the guided session relays the
  existing tokens; it does not weaken the plan→preview→approve→run gate.
- No connector-bundle changes; no changes to `internal/connectors/defs/**`.

## 9. Suggested execution order

Phases 1→4 (A-track foundation) strictly first — everything else composes with cobra
routing and typed config. Then 5→7 (B foundation: events, gate) and 6 (C logging) in any
order or parallel worktrees. Dashboards (10) before wizards (11) — the styles foundation
and event consumption harden on the simpler surface first. Browse/docs (13, 14) after the
wizard wave proves the huh embedding pattern. Traces (12) any time after 3; metrics (17)
after 12. Help-churn (19) and cleanup (22) last, when fixtures change on purpose.
