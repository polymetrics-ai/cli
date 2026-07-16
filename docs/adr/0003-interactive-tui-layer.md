# ADR 0003 — Progress events bus + TTY-gated interactive layer (Bubble Tea v2)

- Status: Accepted (2026-07-16)
- Deciders: user (approved plan; explicitly expanded scope to wizards, browsers, docs
  viewer, and accessibility for humans and agents)
- Context docs: `docs/design/tui-ux-design.md` (full UX design),
  `docs/plans/cli-architecture-v2-improvement-plan.md` (Pillar B),
  `CONTEXT.md` (agentic contract), `internal/cli/agentic_contract_test.go`

## Context

`pm` is agent-first by contract — JSON envelopes on stdout, stderr for diagnostics, no
ANSI anywhere (`safety.SanitizeTerminal`, `internal/safety/safety.go:32`), non-interactive
and deterministic — and human-hostile by accident. Long-running operations (ETL runs, flow
DAGs, certify batches, RLM agent runs) expose no live progress: the flow engine returns a
single `RunResult` and progress is polled from a checkpoint file; certify writes
`progress.json` at the end; the Temporal submit blocks in `run.Get`
(`internal/worker/submit.go:53-60`) with heartbeats invisible to the user. There is no
event stream anywhere. Creation surfaces are worse: flows are hand-authored JSON where DAG
edges are implicit in matching `in`/`out` table names (`internal/flow/dag.go`), schedules
need hand-written cron strings, `pm connectors list` dumps 551 rows through `%+v`, no
command enumerates queryable warehouse tables, and docs render as raw strings. The charm
v2 line (bubbletea v2.0.8, bubbles v2.1.1, lipgloss v2.0.5, huh v2.0.3 with an accessible
mode, glamour v2.0.1, colorprofile) is stable as of 2026-02 on `charm.land` import paths;
gh CLI's accessible-prompter work provides the proven accessibility blueprint.

## Decision

1. **Build a dependency-free progress bus first** (`internal/events`: stdlib +
   `internal/safety` only): a typed `Event` value (kind/scope/run/step/status/counters),
   an `Emitter` carried via `context` with a `Nop` default, and sinks — `NDJSON`
   (sanitized events to **stderr** behind a new `--progress ndjson` flag: live progress for
   agents before any TUI exists), `Chan` (TUI bridge; lifecycle events never dropped,
   progress coalescible), `Throttle`, `Multi`. Instrumentation lands beside the existing
   ledger call sites in `flow.Engine.Run`, in ETL batch `flush()`, in the certify worker
   pool, and as a Temporal `DescribeWorkflowExecution` poller (no workflow code changes).
2. **The TUI is affirmatively gated.** `ui.Detect` enables it only when stdout is a real
   TTY ∧ ¬`--json` ∧ ¬`--plain` ∧ ¬`--no-input` ∧ `PM_NO_TUI`/`CI` unset ∧ `TERM≠dumb`.
   `cli.RunWithOptions` carries the mode; the existing `Run` delegates with plain mode, so
   every existing test exercises the plain path by construction. On the TUI path the
   sanitizer moves into view-string hygiene (every dynamic string sanitized + redacted
   before styling); the plain path is untouched.
3. **Adopt Bubble Tea v2 + bubbles + lipgloss v2** for dashboards and browsers, **huh v2**
   for wizards (embedded as `tea.Model`; accessible mode wired from day one), **glamour v2**
   for the docs pager, **Evertras/bubble-table** for the query grid. Inline mode for run
   commands (final frame persists in scrollback); alt-screen only for browsers/pagers.
4. **Flags are the API; prompts are progressive enhancement.** Wizards prompt only for
   missing inputs, validate with the same code paths the flag door uses (e.g. wizard
   manifests round-trip `flow.ParseManifest`), emit machine artifacts at documented paths,
   and end by printing the scripted equivalent. `--no-input` errors name the exact
   flag/file to provide. New enumerators required by the TUI (`pm query tables`) ship as
   plain/JSON commands first.
5. **Accessibility is a launch requirement, not a follow-up**: huh `WithAccessible` wired
   to `--accessible`/`PM_ACCESSIBLE_PROMPTER`/`ACCESSIBLE`; spinner-disable with static
   status lines; colorprofile degradation honoring NO_COLOR/CLICOLOR/TERM=dumb; a 4-bit
   `accessible_colors` mode; color always paired with glyph + word; minimum-size guard;
   a `pm a11y` help topic. Design language, palette tokens, and per-surface specs live in
   `docs/design/tui-ux-design.md`.
6. **Import direction is law** (CI-checked): business packages may import
   `internal/events` and never `internal/ui`; `internal/ui/**` imports
   events/safety/charm and never business packages; only the command layer branches
   plain-vs-TUI.
7. **No interactive secret entry.** `credentials add` keeps env/stdin intake only; wizard
   assistance is limited to non-secret names and fields.

## Alternatives considered

- **Keep polling files for progress** (status quo): rejected — neither humans nor agents
  can watch a run; every new surface would reinvent polling.
- **TUI without an events bus** (models call engines directly): rejected — couples UI to
  business packages, unusable by agents, and blocks the NDJSON progress feature that
  delivers value with zero new dependencies.
- **Derive events from OTel spans (or vice versa)**: rejected — ties UI availability to
  telemetry configuration and loses context parentage; the layers stay siblings
  (ADR-0004) correlated by run ID.
- **survey/promptui for prompts**: rejected — maintenance-mode libraries, no accessible
  mode, no composition with a live dashboard.
- **Bubble Tea v1**: rejected — greenfield UI code with zero migration burden; v1 is in
  maintenance; huh/glamour/teatest current lines target v2.
- **Alt-screen everywhere**: rejected — run commands must leave a truthful transcript in
  scrollback for humans and CI-attached TTYs.

## Consequences

- (+) One instrumentation pass gives humans live dashboards and agents `--progress ndjson`;
  the flow-creation DAG error class disappears structurally (pickers only offer upstream
  tables).
- (+) The agent contract is preserved by construction: gate defaults, untouched `cli.Run`,
  stderr-only events, contract tests per TUI command.
- (+) Accessibility parity with the best current CLI practice (gh), documented under
  `pm a11y`.
- (−) go.mod grows by the charm v2 suite (+x/term, bubble-table, glamour; ~10–15 modules
  transitively) → contained: charm imports live only under `internal/ui`; phases 5–7 ship
  value with only x/term; the heavy gate is isolated to the dashboard/wizard phases.
- (−) A second rendering regime (styled TTY vs sanitized plain) → mitigated by view-string
  hygiene tests, the import law, and golden teatest frames.
- (−) Windows/terminal variance → inline mode, ASCII glyph fallbacks, colorprofile
  degradation, `PM_NO_TUI` escape hatch.
