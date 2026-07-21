# Phase 408 — Flow/ETL run dashboards

Status: CORRECTION COMPLETE / execute completion false pending Shepherd handoff and independent VERIFY
Issue: #408 `feat(ui): add flow and ETL run dashboards`  
Parent: #397, base `feat/cli-architecture-v2`, parent PR #438  
Worker branch: `feat/408-flow-etl-dashboards`  
Worker directory: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-408-flow-etl-dashboards`

## Scope

Deliver the Phase 10 dashboard slice only:

- Flow and ETL inline run dashboard models.
- Event channel bridge from `internal/events` to Bubble Tea models.
- Ctrl+C cancellation wiring that cancels the engine/run context and waits for a truthful final frame.
- Focused model/bridge/command tests for TTY activation, bypass, parity, cancellation, throttling, layouts, accessibility/plain frames, and view hygiene.
- Narrow help/docs/website parity needed for `pm flow run` / `pm etl run` dashboard behavior.

Out of scope / blocked by instruction:

- #413 shell completion.
- #419 optional log bridge.
- Phase 437 pending-intake requests.
- PR #466 / #437 branch.
- Connector definition changes.
- Reverse ETL execution or new write surfaces.
- Any unapproved dependency, including NTCharts.

## Required context loaded

- `AGENTS.md`
- Issue #408 body and acceptance criteria via `gh issue view 408 --json body --jq .body`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`
- `docs/plans/cli-architecture-v2-improvement-plan.md`
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`
- `docs/design/tui-ux-design.md`
- `docs/design/terminal-ui-research-and-design-system.md`
- `docs/adr/0002-interactive dependencies context via 0002`, `docs/adr/0003-interactive-tui-layer.md`, `docs/adr/0004-observability context`
- `docs/architecture/repo-profile.json`, `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md`, `README.md`

Phase artifacts were absent at worker start; this directory creates the issue-local plan set.

## Resume checkpoint (2026-07-21)

- Resumed at committed plan head `361a6bec0af1ed9cf84d5bdfdd10f16458d9da4d` on `feat/408-flow-etl-dashboards`.
- Adopted all 19 pre-existing dirty entries, including untracked `internal/ui/run/dashboard.go` and `dashboard_test.go`; no reset/stash/clean/restart performed.
- Existing RED and focused GREEN evidence below remain authoritative. EXECUTE continues from that green slice; production behavior still requires live Bubble Tea runner/cancellation reconciliation and broader gates.
- Resume execution decision: `local_critical_path` (sole isolated mutating worker; recursive delegation unavailable by contract).

## Required skills loaded

- `gsd-core` — repo-local GSD adapter; commands used below.
- `bubble-tea-tui-design` — Rule start-here + non-negotiable TTY/plain contract; references loaded:
  - `references/interaction-and-layout.md`
  - `references/charts-and-dashboards.md`
  - `references/testing-and-accessibility.md`
  - `references/inspiration-study.md`
- `golang-how-to` — orchestrator rule: Go task loads relevant secondary skills together.
- `golang-cli` — stdout/stderr, exit-code, signal, flag/help behavior.
- `golang-testing` — table-driven, race, headless frame tests.
- `golang-error-handling` — wrap/return errors once, no log-and-return duplication.
- `golang-security` — sanitize/redact untrusted view strings; no secrets or raw write tools.
- `golang-safety` — nil/resize bounds, no panics on narrow frames.
- `golang-context` — propagate cancellation through run context.
- `golang-concurrency` — owned goroutines, channel lifecycle, throttling without lifecycle loss.
- `golang-documentation` — CLI help/docs/website parity.
- `golang-spf13-cobra` — command wiring changes may touch cobra-backed CLI nodes.
- `caveman` — compact status/handoff only.
- `.pi/skills/go-implementation/SKILL.md` wrapper — absent in live tree (`test -e` => `absent`); no path invented. Available routed Go skills above are implementation authority.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 408 --skip-research
scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run
```

Results:

- `scripts/gsd doctor`: pass.
- `scripts/gsd list`: pass, 69 commands listed.
- `scripts/gsd prompt plan-phase 408 --skip-research`: pass; generated `/tmp/gsd-plan-408.txt` (142 lines).
- `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.

Manual universal-loop fallback is active for implementation because the adapter lacks `programming-loop`; policy source remains `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`. Execution decision for plan cycle: `local_critical_path` (single bounded worker, no subagent tool by design).

## Branch / parent sync

- Started on `feat/408-flow-etl-dashboards` at `5b603788`.
- Parent branch `origin/feat/cli-architecture-v2` at `b77d8f49`.
- Ran `git fetch origin feat/cli-architecture-v2` then `git merge --ff-only origin/feat/cli-architecture-v2`.
- Branch fast-forwarded to parent dispatch head `b77d8f49` before production edits.

## Slice plan

### Slice 0 — Planning checkpoint

- Create issue-local GSD artifacts.
- Record manual-GSD fallback and loaded skills.
- No production edits.
- Commit/push plan checkpoint when clean.

### Slice 1 — RED dashboard/model/bridge contract — RECORDED

Add focused failing tests before production code:

- Dashboard model renders success, failure, cancellation, and final truthful frames.
- Wide/standard/compact/size-guard layouts render without panic and with useful content.
- No-color/ASCII/reduced-motion/accessibility sequential frames preserve word+glyph meaning and no ANSI.
- Control characters and secret-like strings are sanitized/redacted before view output.
- Event bridge throttles/coalesces progress events while delivering lifecycle events in order.
- Ctrl+C path sends cancellation and waits for Done/final frame.
- `flow run` and `etl run` dashboard activation only when stdin+stdout TTY and no bypass.
- Bypass matrix: `--plain`, `--json`, `--no-input`, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`, stdin-piped, stdout-piped all stay plain/noninteractive.
- Machine paths produce no ANSI and keep stdout/stderr/envelope/exit parity.

Expected RED: tests fail to compile or assert missing dashboard package/wiring.

### Slice 2 — Minimal GREEN dashboard package — SUPERSEDED BY SHEPHERD CORRECTION; REAL TEA GREEN RECORDED ABOVE

- Implement `internal/ui/run` (or existing UI package if present) with pure model state and renderer.
- Consume small view DTOs + `events.Event`; do not import business packages into `internal/ui/**`.
- Add event channel bridge with bounded buffering / throttle semantics; lifecycle events never dropped.
- Add cancellation message handling; model exits only after done/final state.
- Keep inline final frame; no alt screen for run dashboards.
- Use existing ADR-approved Charm v2 deps only if already approved/present; stop for any dependency deviation.

### Slice 3 — Command wiring

- Wire `pm flow run` and `pm etl run` to dashboard only on `ui.Detect` eligible mode.
- Plain path remains existing behavior; `cli.Run` default remains plain.
- `--progress ndjson` remains stderr-only if present from #405/#403.
- Ensure cancellation cancels the context used by flow/ETL runner, not only the UI.

### Slice 4 — Parity docs/help

- Update runtime help/docs only for dashboard behavior, no broad help tree churn.
- Update `docs/cli/flow.md`, `docs/cli/etl.md`, and website reference data/docs if current generated parity expects it.
- Add tests or generator checks to prevent drift.

### Slice 5 — Refactor and gates

Run after coherent green slices:

```bash
gofmt -w cmd internal
git diff --check
go test ./internal/ui/... ./internal/cli/... ./internal/flow/... ./internal/app/...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

`go test -race ./...` is phase gate target when feasible; if too slow/blocking, record exact blocker and run focused race tests.

## CLI help/docs/website parity checklist

- [ ] `pm help flow` mentions eligible TTY dashboard and bypass behavior.
- [ ] `pm help etl` mentions eligible TTY dashboard and bypass behavior.
- [ ] `pm flow` bare namespace remains contextual help / exits 0.
- [ ] `pm etl` bare namespace remains contextual help / exits 0.
- [ ] `pm flow run --help` / relevant help output accurate for `--plain`, `--json`, `--no-input`, `--progress ndjson` where supported.
- [ ] `pm etl run --help` / relevant help output accurate for `--plain`, `--json`, `--no-input`, `--progress ndjson` where supported.
- [ ] `docs/cli/flow.md` updated or marked not applicable.
- [ ] `docs/cli/etl.md` updated or marked not applicable.
- [ ] `website/**` updated or marked not applicable with grep evidence.
- [ ] Generated help/manual artifacts/goldens updated or marked not applicable.
- [ ] Invalid actions still return usage errors.
- [ ] JSON/plain/stdout/stderr paths documented and unchanged.

## Shepherd correction slice — real Bubble Tea v2 (2026-07-20)

Authority reconciled before production edits: accepted ADR-0003 decision 3, accepted parent plan Stage 10, and issue #408 authorize only these direct pins: `charm.land/bubbletea/v2@v2.0.8`, `charm.land/bubbles/v2@v2.1.1`, `charm.land/lipgloss/v2@v2.0.5`, and test-only `github.com/charmbracelet/x/exp/teatest/v2@v2.0.0-20260720091843-3eef36eaaa28`. Go-produced transitives are permitted; NTCharts, huh, glamour, beta OTel logs, and every other new direct module remain forbidden.

Correction order:

1. Add a strict compile/runtime RED proving the current dashboard is not a Bubble Tea v2 `tea.Model` and is not driven by `teatest/v2`; capture exact output before production edits.
2. Replace the production custom headless-only session substitute with the smallest inline `tea.Program`: v2 `Init`, deterministic `Update`, `View`; event/cancel/resize/key messages flow through Tea; async wait/cancel work remains in `tea.Cmd`; no alt screen; final truthful frame stays in scrollback.
3. Drive success/failure/cancel and responsive 160x45, 100x30, 80x24, compact, and guard frames through real `teatest/v2`; retain lifecycle delivery, bounded refresh, navigation/help, sanitation/redaction, and plain/JSON/non-TTY bypass behavior.
4. Run focused GREEN/refactor/race gates, vet, build, and full non-race tests as feasible. Do **not** rerun `make verify` or full race during CORRECT. Independent VERIFY owns those gates and the preserved timeout disposition.
5. Commit/push coherent RED, GREEN, and synchronized artifact checkpoints. Do not invoke VERIFY/REVIEW/INTEGRATE or open a sub-PR.

Correction execution decision: `local_critical_path` — exactly one isolated Sol/high correction worker; no subagent tool and no other live worker. Implementation and focused gates are green at `c70ecf64`; `execute_complete=false` until independent VERIFY.

Prior evidence remains immutable: full `go test -race ./...` timed out at 10m; `go test -race -timeout 20m ./internal/cli` timed out without race findings. Prior `make verify` crossed the narrower worker dispatch boundary but used only a temporary local fixture and preserved reverse ETL plan → preview → approval → execute with no credential, remote, production, or persistent write. This is a prior dispatch-boundary deviation, not a fabricated verification failure; CORRECT will not rerun it.

## Safety notes

- No secrets requested, printed, stored, summarized, or invented.
- No credentialed connector checks.
- No reverse ETL execution.
- No external destructive actions or production services.
- No generic shell/HTTP/SQL write surfaces.
- No new dependencies beyond exact ADR-0003 approved modules/versions for this phase; NTCharts is unapproved and forbidden.
- No push to `main`; parent PR merge remains human-gated.
