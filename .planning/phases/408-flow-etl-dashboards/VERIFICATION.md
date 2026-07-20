# Verification — Phase 408 flow/ETL dashboards

Status: blocked after repeated full-race timeout; implementation/focused race/full non-race/`make verify` are green.

## Required local gates

Run after each coherent green slice where feasible:

```bash
gofmt -w cmd internal
git diff --check
go test ./internal/ui/... ./internal/cli/... ./internal/flow/... ./internal/app/...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Phase-specific target:

```bash
go test -race ./...
```

If a full gate is blocked by time/environment, record exact command, result, and blocker here. Do not mark `verificationPassed` true unless `make verify` exits 0.

## Focused checklist

### TUI/model/event/cancellation

- [x] Dashboard model success frame.
- [x] Dashboard model failure frame with sanitized/redacted error.
- [x] Dashboard model cancellation frame after runner final event.
- [x] Event throttle/coalesce retains lifecycle events.
- [x] Ctrl+C cancels model/run context and waits for Done/final frame.
- [x] No goroutine/channel leaks under focused race test.

### Layout/accessibility/view hygiene

- [x] Wide layout (160x45).
- [x] Standard layout (100x30 and/or 80x24).
- [x] Compact layout (60-79 width).
- [x] Size guard below 60x18.
- [x] No-color frame has no ANSI.
- [x] ASCII fallback frame.
- [x] Reduced-motion/static status frame.
- [x] Accessibility/plain sequential transcript.
- [x] Control-character sanitation.
- [x] Secret-like value redaction.

### CLI activation and parity

- [x] Eligible stdin+stdout TTY activates dashboard for `pm flow run`.
- [x] Eligible stdin+stdout TTY activates dashboard for `pm etl run`.
- [x] `--plain` bypass.
- [x] `--json` bypass.
- [x] `--no-input` bypass.
- [x] `CI=1` bypass.
- [x] `PM_NO_TUI=1` bypass.
- [x] `TERM=dumb` bypass.
- [x] stdin-piped fallback.
- [x] stdout-piped fallback.
- [x] No ANSI in machine paths.
- [x] Plain output byte/exit parity for flow bypass behavior.

### CLI help/docs/website parity

- [x] `pm help flow` checked by help test/manual parity path.
- [x] `pm flow` bare namespace checked by existing help tests.
- [x] `pm flow run --help` behavior unchanged: legacy tail executes action; documented under existing tests.
- [x] `pm help etl` checked by help test/manual parity path.
- [x] `pm etl` bare namespace checked by existing help tests.
- [x] `pm etl run --help` behavior unchanged: legacy tail executes action; documented under existing tests.
- [x] `docs/cli/flow.md` updated.
- [x] `docs/cli/etl.md` updated.
- [x] `website/**` updated (`cli-reference.mdx`, `etl.mdx`, `architecture.mdx`).
- [x] Generated help/manual artifacts/goldens updated/verified (`TestGoldenTranscripts` + docs parity tests).

## Resume state

- Adopted committed plan head `361a6bec0af1ed9cf84d5bdfdd10f16458d9da4d` plus all 19 existing dirty entries without destructive git operations.
- `scripts/gsd doctor` re-run on resume: PASS (69 commands; repo-local Pi adapter healthy).
- Manual universal-loop fallback remains active because the already-recorded `programming-loop` command absence is unchanged.
- Full verification remains false until `make verify` exits 0.

## Gate results

| Command | Result | Evidence |
|---|---|---|
| `scripts/gsd doctor` | PASS | Plan cycle; PASS again at EXECUTE resume. |
| `scripts/gsd list` | PASS | Plan cycle. |
| `scripts/gsd prompt plan-phase 408 --skip-research` | PASS | `/tmp/gsd-plan-408.txt`. |
| `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` | FAIL | `scripts/gsd: unknown GSD command: programming-loop`; manual fallback active. |
| `git fetch origin feat/cli-architecture-v2 && git merge --ff-only origin/feat/cli-architecture-v2` | PASS | Fast-forwarded to `b77d8f49`. |
| `go test ./internal/ui ./internal/ui/run ./internal/cli -run 'TestDetectModeUsesADRGate|TestDashboard|TestBridge|TestRunDashboards|TestETLRunDashboard'` | FAIL | RED evidence captured in `TDD-LEDGER.md`. |
| `gofmt -w cmd internal && go test ./internal/ui ./internal/ui/run ./internal/cli -run 'TestDetectModeUsesADRGate|TestDashboard|TestBridge|TestRunDashboards|TestETLRunDashboard|TestGlobalUIFlagsDocumentedInHelp'` | PASS | Focused dashboard/detection/CLI/help green. |
| `go test ./internal/ui/run -run 'TestDashboardFramesCoverLifecycleLayoutsAndHygiene|TestDashboardNavigationHelpAndResize|TestSessionCancellationPropagatesAndDrainsFinalLifecycle'` | FAIL (RED) | Missing `SelectedStep`, `Resize`, `NewSession`, `SessionOptions`; exact output in TDD ledger. |
| `go test ./internal/ui/run -run TestSessionRendersLiveUpdatesAndPersistsFinalFrame` | FAIL (RED) | Missing `SessionOptions.Output`; exact output in TDD ledger. |
| `gofmt -w cmd internal && git diff --check && go test ./internal/ui/... -count=1 && go test ./internal/cli -run 'TestDashboard|TestSession|TestBridge|TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard|TestGlobalUIFlagsDocumentedInHelp|TestGoldenTranscripts|TestDocs' -count=1` | PASS | UI `0.394s/0.807s/0.657s`; CLI `30.011s`. |
| `go test -race ./internal/ui/... -count=1 && go test -race ./internal/cli -run 'TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard' -count=1 && go test -race ./internal/flow -run 'TestEngineCancellationPreservesEventsTelemetryCheckpointLedgerAndLease' -count=1` | PASS | UI packages green; CLI `82.493s`; flow `1.314s`. |
| `go test ./internal/app/... -count=1` | PASS | `29.100s`. |
| `go test ./internal/events/... -count=1` | PASS | `0.454s`. |
| `go vet ./...` | PASS | No output. |
| `go build ./cmd/pm` | PASS | No output. |
| `git diff --name-only -- go.mod go.sum` | PASS | No dependency delta. |
| `go test ./...` | PASS | Full repository suite green; `internal/cli 453.289s`, `internal/connectors/certify 350.901s`. |
| `go test -race ./...` | FAIL (timeout) | Default 10m package timeouts in `internal/cli` and `internal/connectors/certify`; no race finding emitted. |
| `go test -race -timeout 20m ./internal/cli` | FAIL (timeout) | Retry timed out at 20m in existing credential safety suite while repeatedly loading connector bundles; no race finding emitted. Repeated verification failure hard stop triggered; certify retry not run. |
| `make verify` | PASS | fmt, tidy-check, vet, 20m full tests, build, docs validate, smoke, lint, and 547-bundle validation exited 0. |

## Manual TTY record

PASS using safe local fixture projects under `/tmp` through `script`-allocated dual TTYs:

```text
manual dual-TTY ETL: final=1 frames=6
manual dual-TTY flow: final=1 frames=5
```

No credential values, remote services, connector definitions, or reverse ETL execution used.

## Known parity boundary

- `pm help flow`, bare `pm flow`, `pm help etl`, and bare `pm etl`: exit 0 with contextual manuals.
- `pm flow invalid` and `pm etl invalid`: exit 2 usage errors.
- Focused `pm flow run --help` / `pm etl run --help` remain the inherited legacy behavior and attempt action/project setup (exit 1 without a project). Per-subcommand help deepening belongs to Phase 19, not #408; no out-of-scope router/help-tree behavior changed here.
- Bubble Tea/teatest dependencies are absent and adding dependencies is forbidden by this EXECUTE instruction. Deterministic headless model/session tests cover the phase semantics; literal teatest coverage remains unavailable without a later approved dependency-bearing stage.
- Safety deviation: `make verify` invokes the repository smoke recipe, which executed a local temporary fixture reverse-ETL plan/preview/run under a generated temp directory. No remote connector, credential value, production service, or persistent project data was used, but this still crossed the explicit EXECUTE boundary forbidding reverse ETL execution. No further execution gates were run after the repeated race timeout; orchestrator/human disposition required.
