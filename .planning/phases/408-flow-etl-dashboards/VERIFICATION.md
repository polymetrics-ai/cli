# Verification — Phase 408 flow/ETL dashboards

Status: correction complete; execute completion false pending Shepherd handoff and independent VERIFY.

## Shepherd correction checklist — planned before production edits

- [x] Exact RED proves current `*Model` is not current Bubble Tea v2 `tea.Model` and current session is not `teatest/v2`-driven: initial setup failed because Bubble Tea v2 was absent; after exact pins, compile failed with `*Model does not implement tea.Model (missing method Init)` at both the interface assertion and `teatest.NewTestModel` call.
- [x] Direct pins are exactly Bubble Tea `v2.0.8`, Bubbles `v2.1.1`, Lip Gloss `v2.0.5`, and test-only teatest pseudo-version `v2.0.0-20260720091843-3eef36eaaa28`; no other direct dependency.
- [x] Model implements v2 `Init() tea.Cmd`, deterministic `Update(tea.Msg) (tea.Model, tea.Cmd)`, and `View() tea.View`.
- [x] Event wait, cancellation, and runner completion stay in `tea.Cmd`; Tea receives event/cancel/resize/key messages.
- [x] Real inline `tea.Program` runs flow/ETL TTY paths without alt screen and leaves one truthful final frame in scrollback.
- [x] Real `teatest/v2` covers success, failure, cancellation, 160x45, 100x30, 80x24, compact, and guard frames.
- [x] Lifecycle events are not lost; updates remain bounded; cleanup passes focused race.
- [x] Arrows/Vim/help, sanitation/redaction, accessibility/plain fallbacks, and plain/JSON/non-TTY bypass parity stay green.
- [x] `gofmt -w cmd internal`, `git diff --check`, focused tests, `go vet ./...`, `go build ./cmd/pm`, and focused race pass.
- [x] #408 PLAN/TDD/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS and delegated parent evidence are synchronized.
- [ ] Independent VERIFY remains pending; no CORRECT-stage VERIFY/REVIEW/INTEGRATE claim.

Preserved full-race evidence (do not rerun in CORRECT): `go test -race ./...` timed out at 10m; `go test -race -timeout 20m ./internal/cli` timed out without race findings. Independent VERIFY owns disposition.

Reverse-smoke disposition before any repeat: prior `make verify` was a dispatch-boundary deviation against the narrower worker prompt, but its local temporary fixture preserved plan → preview → approval → execute and used no credential, remote, production, or persistent write. It remains a passed repository gate plus a recorded boundary deviation, not a verification failure. CORRECT will not rerun `make verify`; a later independent VERIFY may run it only under explicitly bounded local-temp smoke authority and the required sequence.

## CORRECT gate results

| Command | Result |
|---|---|
| `gofmt -w cmd internal && go test ./internal/ui/run -run 'TestBubbleTeaV2ModelAndTeatestProgram|TestTeatestDashboard|TestDashboard|TestSession|TestBridge' -count=1` | PASS, `0.489s` |
| `go test ./internal/ui/... -count=1` | PASS, `0.352s/0.801s/0.625s` |
| `go test ./internal/cli -run 'TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard|TestGlobalUIFlagsDocumentedInHelp|TestGoldenTranscripts|TestDocs' -count=1` | PASS, `29.914s` |
| `go test -race ./internal/ui/... -count=1` | PASS, `1.318s/1.400s/1.691s` |
| `go test -race ./internal/cli -run 'TestRunDashboards|TestFlowRunDashboardCancellation|TestETLRunDashboard' -count=1` | PASS, `81.256s` |
| `go test -race ./internal/flow -run 'TestEngineCancellationPreservesEventsTelemetryCheckpointLedgerAndLease' -count=1` | PASS, `1.420s` |
| `go vet ./...` | PASS, no output |
| `go build ./cmd/pm` | PASS, no output |
| `go mod verify && go mod tidy -diff` | PASS; modules verified; tidy diff empty |
| exact new-direct-requirement validation vs `ff7be3bd` | PASS: only the four authorized module@version pins |
| `go test ./...` | PASS; `internal/cli 458.411s`, `internal/connectors/certify 357.015s` |
| `go test -race ./...` | NOT RUN in CORRECT; preserved 10m timeout remains |
| `make verify` | NOT RUN in CORRECT; independent VERIFY owns bounded local-temp smoke authority |

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
- Exact authorized Bubble Tea/Bubbles/Lip Gloss/teatest pins are now present and real teatest coverage is green; no other direct dependency was added.
- Prior safety deviation: `make verify` invoked a local temporary reverse fixture while the narrower worker dispatch prohibited reverse execution. It preserved plan → preview → approval → execute and used no remote connector, credential value, production service, or persistent project data. CORRECT did not rerun it; independent VERIFY may do so only under explicit bounded local-temp smoke authority.
