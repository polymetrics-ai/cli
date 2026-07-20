# TDD Ledger — Phase 408 flow/ETL dashboards

Issue: #408  
Mode: manual universal-loop fallback after `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`.

## Loaded skills

- `gsd-core`
- `bubble-tea-tui-design` + references: interaction/layout, charts/dashboards, testing/accessibility, inspiration study
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-spf13-cobra`
- `caveman` for final handoff only
- `.pi/skills/go-implementation/SKILL.md` checked and absent; used available routed Go skills without inventing the wrapper

## Shepherd correction cycle — planned before production edits

Loaded skills remain: `gsd-core`; `bubble-tea-tui-design` interaction/layout, dashboard, and testing/accessibility references; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`; `caveman` handoff-only.

Rules applied: Bubble Tea skill non-negotiable model/command/TTY contract and Bubble Tea v2 mechanics; `golang-testing` rules 1, 3, 5, and 10; `golang-concurrency` rules 1, 4, and 7; `golang-context` rules 1 and 5; `golang-security` trust-boundary and secret/logging rules; `golang-safety` nil/bounds/resource rules; `golang-error-handling` rules 1, 2, and 7; `golang-cli` stdout/stderr and signal rules; `golang-spf13-cobra` fresh-tree and injected-writer rules.

Strict RED to capture before correction production code/dependency edits:

```bash
go test ./internal/ui/run -run '^TestBubbleTeaV2ModelAndTeatestProgram$' -count=1
```

Expected: current head fails because `*run.Model` does not implement current v2 `tea.Model`, no real `tea.Program` drives the session, and `teatest/v2` is not in the module. GREEN must use exact authorized pins and real teatest programs for success/failure/cancel and responsive frames.

Correction status: RED captured; execute completion false. Decision: `local_critical_path`.

### Shepherd RED evidence

```bash
go test ./internal/ui/run -run '^TestBubbleTeaV2ModelAndTeatestProgram$' -count=1
```

Result: FAIL at pushed baseline plus plan checkpoint, before dependency or production edits.

```text
# polymetrics.ai/internal/ui/run
internal/ui/run/bubbletea_v2_test.go:7:2: no required module provides package charm.land/bubbletea/v2; to add it:
	go get charm.land/bubbletea/v2
FAIL	polymetrics.ai/internal/ui/run [setup failed]
FAIL
```

The strict test imports current v2 `tea.Model`, asserts `var _ tea.Model = (*Model)(nil)`, and instantiates `teatest.NewTestModel`; current production has neither dependency nor interface/program implementation. Next RED after adding only authorized pins must expose the incompatible `View() string`/missing `Init`/`Update` contract before production edits.

## RED plan

Before production edits, capture failing tests/validation for:

1. `internal/ui` dashboard model frames:
   - success final frame
   - failure final frame with redacted/sanitized error
   - cancellation final frame after Done
   - wide/standard/compact/guard layouts
   - no-color/ASCII/reduced-motion/accessibility frames
2. Event channel bridge:
   - progress throttling/coalescing
   - lifecycle events not dropped
   - channel close sends final Done/error state
3. Cancellation:
   - `ctrl+c` cancels runner context
   - model waits for final event before quitting
4. CLI activation matrix for `flow run` and `etl run`:
   - eligible dual-TTY activates dashboard
   - `--plain`, `--json`, `--no-input`, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`, stdin-piped, stdout-piped bypass dashboard/prompt paths
   - no ANSI on machine paths
   - stdout/stderr/exit parity for plain existing behavior
5. Help/docs parity:
   - `pm help flow`, `pm help etl`, bare `pm flow`, bare `pm etl`, and command help reflect behavior or remain unchanged when not applicable

## Evidence log

| Cycle | Type | Command | Result | Notes |
|---|---|---|---|---|
| plan | GSD preflight | `scripts/gsd doctor` | PASS | Adapter healthy. |
| plan | GSD preflight | `scripts/gsd list` | PASS | 69 commands listed. |
| plan | GSD plan prompt | `scripts/gsd prompt plan-phase 408 --skip-research` | PASS | Wrote `/tmp/gsd-plan-408.txt`. |
| plan | GSD programming loop | `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` | FAIL | `scripts/gsd: unknown GSD command: programming-loop`; manual universal-loop fallback recorded. |
| plan | parent sync | `git fetch origin feat/cli-architecture-v2 && git merge --ff-only origin/feat/cli-architecture-v2` | PASS | Branch fast-forwarded to `b77d8f49` before production edits. |

## RED evidence

Captured before production edits.

```bash
go test ./internal/ui ./internal/ui/run ./internal/cli -run 'TestDetectModeUsesADRGate|TestDashboard|TestBridge|TestRunDashboards|TestETLRunDashboard'
```

Result: FAIL as expected.

```text
# polymetrics.ai/internal/ui [polymetrics.ai/internal/ui.test]
internal/ui/detect_test.go:13:24: unknown field StdinTTY in struct literal of type DetectOptions
internal/ui/detect_test.go:18:24: unknown field StdinTTY in struct literal of type DetectOptions
internal/ui/detect_test.go:23:24: unknown field StdinTTY in struct literal of type DetectOptions
internal/ui/detect_test.go:97:29: unknown field StdinTTY in struct literal of type DetectOptions
FAIL	polymetrics.ai/internal/ui [build failed]
# polymetrics.ai/internal/ui/run [polymetrics.ai/internal/ui/run.test]
internal/ui/run/dashboard_test.go:13:11: undefined: NewModel
internal/ui/run/dashboard_test.go:13:20: undefined: Config
internal/ui/run/dashboard_test.go:20:12: undefined: Step
internal/ui/run/dashboard_test.go:63:8: undefined: Config
internal/ui/run/dashboard_test.go:68:10: undefined: Config
internal/ui/run/dashboard_test.go:68:98: undefined: Step
internal/ui/run/dashboard_test.go:73:10: undefined: Config
internal/ui/run/dashboard_test.go:73:97: undefined: Step
internal/ui/run/dashboard_test.go:78:10: undefined: Config
internal/ui/run/dashboard_test.go:78:97: undefined: Step
internal/ui/run/dashboard_test.go:78:97: too many errors
FAIL	polymetrics.ai/internal/ui/run [build failed]
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/ui_options_test.go:228:30: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:249:84: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:250:85: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:251:49: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:252:56: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:253:56: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:254:58: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:255:59: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:269:103: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:285:30: unknown field StdinIsTerminal in struct literal of type RunOptions
internal/cli/ui_options_test.go:285:30: too many errors
FAIL	polymetrics.ai/internal/cli [build failed]
FAIL
```

RED coverage introduced:

- Dual-TTY detection requires stdin and stdout.
- Flow/ETL TTY dashboard activation and bypass matrix.
- Dashboard model success/failure/cancel, layout, accessibility/ASCII/no-color, sanitation/redaction.
- Bridge throttling without lifecycle loss.

## GREEN evidence

Focused dashboard slice green:

```bash
gofmt -w cmd internal && go test ./internal/ui ./internal/ui/run ./internal/cli -run 'TestDetectModeUsesADRGate|TestDashboard|TestBridge|TestRunDashboards|TestETLRunDashboard|TestGlobalUIFlagsDocumentedInHelp'
```

Result: PASS.

```text
ok  	polymetrics.ai/internal/ui	0.440s
ok  	polymetrics.ai/internal/ui/run	0.668s
ok  	polymetrics.ai/internal/cli	7.942s
```

Implemented minimal green:

- Added stdin+stdout TTY gate support via `DetectOptions.StdinTTY` and `RunOptions.StdinIsTerminal`.
- `cmd/pm` now uses `RunWithOptions(... ModeAuto)`; `cli.Run` remains plain for agent/certify seams.
- Added `internal/ui/run` deterministic dashboard model and event bridge.
- Wired `flow run` and `etl run` TTY path to render final inline dashboard frames; plain/JSON/no-input/bypass paths remain existing output.
- Updated runtime help, `docs/cli/{flow,etl,config}.md`, and website docs for dashboard/bypass behavior.

## Resume RED — live session/navigation/rate hardening

Captured before additional production edits after adopting the focused GREEN slice.

```bash
go test ./internal/ui/run -run 'TestDashboardFramesCoverLifecycleLayoutsAndHygiene|TestDashboardNavigationHelpAndResize|TestSessionCancellationPropagatesAndDrainsFinalLifecycle'
```

Result: FAIL as expected.

```text
# polymetrics.ai/internal/ui/run [polymetrics.ai/internal/ui/run.test]
internal/ui/run/dashboard_test.go:126:18: model.SelectedStep undefined (type *Model has no field or method SelectedStep)
internal/ui/run/dashboard_test.go:130:18: model.SelectedStep undefined (type *Model has no field or method SelectedStep)
internal/ui/run/dashboard_test.go:134:18: model.SelectedStep undefined (type *Model has no field or method SelectedStep)
internal/ui/run/dashboard_test.go:139:18: model.SelectedStep undefined (type *Model has no field or method SelectedStep)
internal/ui/run/dashboard_test.go:143:18: model.SelectedStep undefined (type *Model has no field or method SelectedStep)
internal/ui/run/dashboard_test.go:155:8: model.Resize undefined (type *Model has no field or method Resize)
internal/ui/run/dashboard_test.go:193:13: undefined: NewSession
internal/ui/run/dashboard_test.go:193:32: undefined: SessionOptions
FAIL polymetrics.ai/internal/ui/run [build failed]
FAIL
```

Contract added: arrows/Vim-equivalent selection and help, one-layer `esc`, resize guard, exact `records/s` rate, and parent-context cancellation that drains terminal lifecycle events before returning.

Additional live-render RED captured before renderer production edits:

```bash
go test ./internal/ui/run -run TestSessionRendersLiveUpdatesAndPersistsFinalFrame
```

Result: FAIL as expected.

```text
# polymetrics.ai/internal/ui/run [polymetrics.ai/internal/ui/run.test]
internal/ui/run/dashboard_test.go:201:3: unknown field Output in struct literal of type SessionOptions
FAIL polymetrics.ai/internal/ui/run [build failed]
FAIL
```

## Resume GREEN evidence

```bash
gofmt -w cmd internal && go test ./internal/ui/run -run 'TestSession|TestDashboard|TestBridge' -count=1 && go test ./internal/cli -run 'TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard' -count=1
```

Result: PASS.

```text
ok  polymetrics.ai/internal/ui/run  0.461s
ok  polymetrics.ai/internal/cli     7.623s
```

Minimal resumed GREEN:

- event-driven session drains throttled progress and every lifecycle event before returning;
- parent/SIGINT cancellation propagates to the flow/ETL run context and preserves the terminal frame;
- live inline refreshes use the existing dependency set and leave the final frame in scrollback;
- responsive initial dimensions, rate units, navigation/help state, sanitation/redaction, and plain/JSON bypasses remain deterministic.

## REFACTOR evidence

```bash
gofmt -w cmd internal && git diff --check && go test ./internal/ui/... -count=1 && go test ./internal/cli -run 'TestDashboard|TestSession|TestBridge|TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard|TestGlobalUIFlagsDocumentedInHelp|TestGoldenTranscripts|TestDocs' -count=1 && go test -race ./internal/ui/... -count=1 && go test -race ./internal/cli -run 'TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard' -count=1 && go test -race ./internal/flow -run 'TestEngineCancellationPreservesEventsTelemetryCheckpointLedgerAndLease' -count=1
```

Result: PASS. Focused model/CLI/race coverage green; full repository gates remain tracked in `VERIFICATION.md`.

Dependency note: no `go.mod`/`go.sum` delta. Bubble Tea/teatest are absent from the live module and the EXECUTE instruction forbids new dependencies, so coverage uses deterministic headless semantic/model tests rather than adding teatest.

Broader REFACTOR gates: `go test ./...` PASS; `make verify` PASS. Full `go test -race ./...` timed out in `internal/cli` and `internal/connectors/certify` at the default 10m package limit; focused issue races passed. A targeted `go test -race -timeout 20m ./internal/cli` retry also timed out with no race finding, triggering the repeated-verification-failure hard stop. Exact details are in `VERIFICATION.md`.
