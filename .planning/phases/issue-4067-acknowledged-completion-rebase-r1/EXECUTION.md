# #4067 execution record

**Status:** R3 GREEN, deterministic core/race, #4046/R7/R8, affected-package, repository, website, existing-connector limitation, and inline GSD verification/review are locally complete. The loop's unrelated warehouse-architecture documentation commit is restored in the immediately following scope-restoration commit. Current no-mistakes help has no safe existing-#4059 delivery route, so only local loop 4/5 remains here; push/PR/CI and exact-head CI remain pending Firstmate direction. No external-check, certification, or merge claim is made here.

## R3 planning and behavioral RED

- The live #4067 body now carries the Sol r3 all-seven two-page typed-conflict contract and the separate existing-connector verification gate. The complete r3 report and updated loop-4 directive were read before local edits.
- Planning/TDD evidence was committed in `0d448ea10` after current #4067/#4059 custody verification, full required-skill loading, GSD adapter/command resolution, and the documented named-phase inline/manual fallback. No production file changed in that checkpoint.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedPageThenStaleSecondPageFinalizesLosingRunForAllModes$' -v` exited `1` as required. In every canonical mode, a real JSON-store loser durably acknowledged page one; a separately opened App persisted unrelated state and completed a real winner over the same target stream; the loser then applied page two exactly once and received `errTransportStreamStateConflict` from the existing per-stream CAS.
- Before asserting the symptom, each subtest proved typed error identity, page-one acknowledgement, winner checkpoint/run preservation, unrelated stream/checkpoint/run preservation, exactly two loser applies, and one winner apply. The rejected behavior then returned `Run{}` and reopened the losing run as durable `running`.
- This RED checkpoint changes only `internal/app/transport_dispatch_test.go` and this GSD evidence. `internal/app/app.go` and `internal/app/transport_dispatch.go` remain untouched until the RED commit exists.

## R3 GREEN — typed-conflict finalization ordering

- `failAcknowledgedTransportRun` now checks `errors.Is(runErr, errTransportStreamStateConflict)` first and delegates to the existing #4046 `failRun` terminalizer before consulting `result.PendingStreamState` or an old in-memory page-one witness.
- The change is intentionally one branch in `internal/app/app.go`. Ordinary acknowledged cancellation/source-error handling still enters the r2 exact-stream/running-run finalizer; checkpoint CAS, source identity, destination application, stream winner state, and registration are untouched.
- The exact C12 selector now exits `0` in all seven modes. The typed conflict remains detectable, page one and the real winner are preserved, unrelated stream/checkpoint/run data survives, the loser applies exactly two pages without replay, and the returned failed loser equals the reopened durable failed record.

## R3 focused and repository-local validation

- Deterministic core: `go test -count=3 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedPageThenStaleSecondPageFinalizesLosingRunForAllModes$'` exits `0` in 27.280s. The same selector under `go test -race -count=3 -timeout 20m` exits `0` in 225.052s; an additional `-race -count=1 -v` pass exits `0` in 84.566s. The macOS linker emitted its known malformed `LC_DYSYMTAB` warning only; Go emitted no race report.
- Full transport regression: `go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransport.*|TestFailRunTransport.*|TestFailRunRetainsRevisionGuardWithoutTransportConflict)$' -v` exits `0` in 82.164s. It covers #4046 stale writers, multi-page checkpoints, r2 completion/failure/missing-target/persistence truth, cancellation, R7/R8 resume identity, interim checkpoints, changed/removed/terminal targets, and no replay. Full `go test -count=1 -timeout 20m ./internal/app` exits `0` in 165.199s.
- Existing connector gate: `go build ./cmd/pm` passes. The real binary's read-only `connectors inspect github --json` and `connectors inspect postgres --json` each report source and destination `unsupported`, `certified: false`, and `COMMUNITY BUILD, UNCERTIFIED`. This is the exact production-registration limitation; no GitHub/PostgreSQL transport-role or GitHub↔DuckDB/Parquet round-trip claim is made, and #4015 wiring remains out of scope.
- Current GitHub definition/hook/preflight evidence passes: `go test ./internal/connectors/hooks/github`; selected `internal/connectors/engine` embedded-bundle/rate-limit tests; selected `internal/connectors/commandrunner` GitHub and `TestEveryImplementedCommandPassesRuntimePreflight`; `go test ./internal/connectors -run '^TestSyncTransport'`; `go test ./internal/synctransport`; selected connector/GitHub CLI transport inspection/direct-read tests; and `node --test scripts/tests/github-live-proof-sweep.test.mjs` (7/7). Those tests are local fixture/harness proof, not provider certification.
- Credentialed bounded smoke: not run. This custody received neither an approved GitHub credential/name nor a sanctioned secret-channel invocation. No environment scan, credential creation, token copy, provider mutation, or substitute public request was attempted. The controlling directive makes that smoke conditional; its absence is recorded rather than inferred away.
- Quality/generator/GSD gates pass: `scripts/verify-gsd-workflow 30b2fb4aeb121641b6158903fe1d3b54668599a6 HEAD`, `go run ./cmd/connectorgen certification-matrix --check`, `pnpm --dir website run gen:website-data` plus clean diff, `go vet ./...`, `make lint`, and the individual tidy, agent-contract, connectorgen validate/surface-sync, connector-boundary, connector-runtime-preflight, release-workflow, GitHub-parity-artifacts, connector-canon, and docs checks.
- `make smoke-no-build` remains intentionally unrun because it mutates a local warehouse, prohibited by this correction; `make verify` remains intentionally not invoked as one command under the repository's per-command timeout guidance.

## R2 planning and behavioral RED

- Planning/TDD continuation committed in `c08a5861f` after the complete Sol r2 handoff, live #4067 amendment/readback, GSD prompt resolution, and the documented named-phase manual fallback. No production file changed in that checkpoint.
- `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes$'` exited `1` as required. All seven modes persisted the acknowledged target checkpoint and an unrelated stream/checkpoint/run and made exactly one destination apply, but cancellation after release returned `Run{}` and reopened the original run as `running`.
- `go test -json -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedFailurePreservesSourceError|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes)$'` exited `1` as required. The source sentinel remained detectable but took the same zero-run/durable-running path; all seven missing-target cases left persisted state unchanged but returned plain `run "..." not found` without `errors.Is(err, errStateRevisionConflict)`.
- The RED commit contains only `internal/app/transport_dispatch_test.go` plus this r2 test evidence. `internal/app/app.go` and `internal/app/transport_dispatch.go` remain untouched until that commit exists.

## R2 GREEN — acknowledged-error finalization and typed missing target

- `runTransportETL` now returns a result witness only after its real committed checkpoint callback has completed. Pre-ack errors retain the zero result and still take ordinary `failRun`.
- The declared-transport error branch invokes a dedicated finalizer only with that witness. Its latest-state update requires the exact acknowledged stream state and an exact `running` target run, then changes only that run to `failed`; it neither refreshes broad state nor retries checkpoint/source/destination work.
- `completeRunWithAcknowledgedTransportState` now makes a missing run in an acknowledged rebase a typed `errStateRevisionConflict`. Ordinary missing-run behavior remains unchanged.
- `go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes|TestRunETLTransportAcknowledgedFailurePreservesSourceError|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes)$' -v` — exit `0`. It covers all seven cancellation and missing-run modes plus the representative source-error path.
- `go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes|TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges|TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes|TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome|TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes|TestRunETLTransportAcknowledgedFailurePreservesSourceError|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes)$' -v` — exit `0`. This jointly reruns the prior completion guard, cancellation, all seven r2 interleavings, missing-target typed chain, and existing completion outcome cases.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome$' -v` — exit `0`. Definite pre-commit failure returns `Run{}` and preserves reopened state; committed unlock and indeterminate directory-sync outcomes return a matching reopened failed run while retaining both the original source sentinel and `CommitOutcomeError`.
- The only production files in this checkpoint are `internal/app/app.go` and `internal/app/transport_dispatch.go`; the corresponding focused test is `internal/app/transport_dispatch_test.go`. No generator, provider, credential, network, warehouse, container, or external service was used.

## R2 focused and heavy local validation

- `go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes|TestRunETLTransportAcknowledgedFailurePreservesSourceError|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes)$' -v` — exit `0`.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome$' -v` and `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges$' -v` — exit `0`.
- The combined all-seven completion/cancellation/error/missing-target suite, the R7/R8 selector, the #4046 repeated/reopen/cancellation selector, and the #4046 all-seven stale-writer selector each exit `0`.
- `go test -race -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes|TestRunETLTransportAcknowledgedFailurePreservesSourceError|TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes|TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome|TestRunETLTransportStaleWriterFinalizesLosingRun)$'` — exit `0`; the macOS linker emitted its known malformed `LC_DYSYMTAB` warning, and Go reported no race.
- `go test -timeout 20m ./internal/app`, `go test -timeout 20m ./internal/synctransport`, `go test -timeout 20m ./internal/connectors/...`, the focused `internal/cli` command-surface selector, `go vet ./...`, and `go build ./cmd/pm` — each exit `0`.
- Individual repository gates `make tidy-check`, `make lint`, `make docs-check`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make release-workflow-check`, `make github-parity-artifacts-check`, `make connector-canon-check`, and `make connector-runtime-preflight` — each exit `0`.
- In a disposable detached worktree, `pnpm run gen:website-data && git diff --exit-code && git status --short` exits `0`. The detached worktree had no installed packages; an offline frozen install stopped at `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`, so active-worktree read-only `pnpm run typecheck` and `pnpm run test:scripts` supplied the package checks and both exit `0`. The disposable worktree was then removed.
- `make smoke-no-build` is deliberately not run: it writes a local warehouse and this correction's controlling handoff forbids warehouse mutation. `make verify` is likewise not invoked as one command because the repository instructions require individual gates under the command-timeout environment.
- Manual `verify-work` and standard-depth `code-review` use the documented named-phase fallback: `gsd-sdk` reports `phase_found: false` for this nonnumeric issue phase and custody forbids lifecycle-role spawning. The current R2 records and review disposition are in `VERIFICATION.md`, `UAT.md`, and `REVIEW.md`.

## R2 no-mistakes correction loop 3/5 — passed locally, scope restored

- Fresh run `01KZS4Y49CT5CBPZ6SEGWR5YWT` started at committed `542cf9fe1c918c60fe77a31b625aa4b9d0155862` without `--yes` and with `--skip=push,pr,ci`, because `axi run --help` exposes no safe route that updates only the required existing stacked PR #4059.
- Its source review completed low risk with no finding. Its test phase passed the real JSON-store C8/C9/C10/C11 suite, the #4046/R7/R8 selectors, and the focused race selector; it wrote only an external temporary acceptance transcript permitted by the gate. Documentation and lint also completed with no unresolved finding.
- The document phase nevertheless committed `c83d148f18b03550f64c1cc516ef28506ca0035d`, changing `docs/architecture/github-postgres-warehouse-certification.md`. That warehouse-architecture rewrite is outside #4067 and is restored immediately after required `no-mistakes axi sync --recover`; it is not retained in the final branch diff.
- The terminal result is `passed`. No push, PR creation/update, CI polling, retarget, close, or merge occurred. With no safe existing-PR delivery mechanism in the current tool, delivery stops here for Firstmate direction.

## Historical pre-r2 candidate record (not evidence for this correction)

The remaining entries record the earlier #4067 candidate and its two completed no-mistakes runs. They are retained for audit history only. They do not establish local validation, PR state, CI, generated-artifact state, or review coverage for the r2 implementation above.

## TDD gate

The first execution action is a behavioral RED in `internal/app` against the real persisted JSON state path. It must demonstrate the Sol F1 durable `running` leak and zero/non-terminal returned run after an acknowledged checkpoint and an unrelated post-checkpoint writer. No `internal/` production file may change until that RED command and the test-only commit are recorded below.

## Planned checkpoints

| Checkpoint | Required evidence | Status |
|---|---|---|
| Planning | #4067 issue/readback, manual GSD context/plan/TDD ledger, clean rejected baseline custody | Complete before this commit |
| RED | Focused test, non-zero exit due to durable symptom, test-only commit | Committed in `5db500fad` |
| GREEN | Minimal completion-boundary implementation and same test passing | Committed in `d16767e47` |
| Focused expansion | all modes, reopen, cancellation, fail-closed eligibility, race, #4046/R7/R8 | Complete in `ce15f07e1` |
| Generated remediation | canonical generator commands and candidate-owned diff only | Complete in `291c3449c` |
| Heavy validation | only after required user-facing window notification | Complete; commands recorded in `VERIFICATION.md` |
| Review/no-mistakes/CI | all findings dispositioned; fresh 0/5 no-mistakes; exact-head CI | Review complete; no-mistakes/CI pending |

## RED — all-seven-mode acknowledged-completion leak

- Command: `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes$'`
- Exit: `1` (expected RED)
- Fixture: a source-executor test seam pauses only after the real checkpoint callback has returned. A second App then persists unrelated stream/checkpoint/run data; release lets the original source report exhaustion and reach ordinary `completeRun`.
- Result: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture` each retained the acknowledged target stream and unrelated writer first, then failed with `RunETL returned zero run`, durable target status `running`, and a zero terminal timestamp. The test also required `errors.Is(runErr, errStateRevisionConflict)`.
- Boundary: this checkpoint includes only `internal/app/transport_dispatch_test.go` and planning evidence. `internal/app/app.go` remains untouched until the RED commit is made.

## GREEN — strict acknowledged-target completion

- Command: `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes$' -v`
- Exit: `0`
- Production change: the closed transport `RunETL` branch captures the exact stream state persisted by its own acknowledged checkpoint. On a later whole-state revision mismatch, `completeRun` may use latest locked state only when that exact target stream survives and the target run remains `running`.
- Result: every canonical mode returned and durably reopened the matching `completed` run. The target checkpoint's connection, stream, generation, timestamp, and opaque checkpoint stayed unchanged; only its own final run metadata advanced. The unrelated stream/checkpoint/run survived unchanged.
- State-store outcome: the new rebase path returns a terminal identity only after a successful or may-have-committed terminal write; definite non-commit still returns `Run{}`. Focused outcome proof follows below.

## Focused expansion — completed before broader validation

- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges$' -v` — exit `0`; changed, missing, and terminal target cases return zero with a detectable `errStateRevisionConflict` and do not overwrite current state.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes$' -v` — exit `0`; cancellation after the real checkpoint callback is observed by every canonical-mode source fixture, preserves the checkpoint, and durably fails the run with `context.Canceled` detectable.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome$' -v` — exit `0`; definite pre-commit returns `Run{}`/durable `running`, while committed and indeterminate outcomes return a matching terminal record plus commit-outcome error.
- `go test -count=3 -timeout 20m ./internal/app -run '^(TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes|TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges|TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes|TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome)$'` — exit `0`.
- Same focused set under `go test -race -count=3 -timeout 20m` — exit `0`; the macOS linker emitted its known malformed `LC_DYSYMTAB` warning, but Go reported no race or test failure.
- Exact #4046/R7/R8 focused regression command — exit `0`; source identity, target-entry CAS, stale-writer conflict finalization, cancellation ordering, and state-store outcome protections remain unchanged.

## Generated-artifact remediation — completed by canonical generators

- `cd website && node scripts/gen-docs-data.mjs` and then `pnpm run gen:website-data` — canonical website output refreshed. The full generator left exactly `website/lib/docs.generated.ts` modified; its only content change is the candidate-owned `Connector transport eligibility` section already authored in `website/content/docs/agent-guide.mdx`.
- `go run ./cmd/connectorgen certification-matrix` followed by `go run ./cmd/connectorgen certification-matrix --check` — exit `0`. The generator changed only `internal/connectors/certifications/flow-matrix.json`, refreshing `internal/cli/cli.go:594 → :600` for ETL and `:1723 → :1729` for reverse ETL, exactly matching Sol F3.
- No generated file was hand-edited and the combined generator status contains no additional paths.

## First no-mistakes delivery attempt — dispositioned before final stacked delivery

- Fresh run `01KZRPD9TSDDBG4F39VENDW9N4` started at `6c2a96d5c` without `--yes`, completed its review, focused tests (including the all-seven/race/#4046/R7/R8 selectors), documentation, lint, push, PR, and CI phases, and returned `checks-passed` at `a17d8db98532c6b2569403f6fec30410acf7104b`.
- Its document phase added a directly relevant explanatory comment at `completeAcknowledgedTransportRun`, plus an unrelated `docs/architecture/github-postgres-warehouse-certification.md` warehouse rewrite. The comment is retained; the warehouse rewrite is removed in this scope-restoration commit because #4067 authorizes only the completion boundary, focused tests, phase evidence, and the two required generated artifacts.
- Despite the explicit stacked-only intent, the tool opened #4068 against an incorrect delivery lane. It was confirmed open, unmerged, and duplicate of #4059, then closed unmerged after `checks-passed`; #4059 remains the only open PR for this branch. The next no-mistakes loop skips push/PR/CI solely to avoid recreating that unsupported base-target transition. It remains within the 2/5 correction budget; delivery after local validation will use ordinary push, #4059 update, and #4059 exact-head CI only.

## Second no-mistakes scope-restoration validation — passed locally

- Fresh run `01KZRTCV26XT15R2YCV6Y4BGMP` started at `f51145c9b` without `--yes` and with `--skip=push,pr,ci`, the narrowly justified delivery fallback above.
- Review, the focused persisted-state/race transport matrix, document, and lint completed with no finding. The document phase preserved one semantically equivalent Markdown hard-break normalization in existing #3864 planning evidence and no production or behavioral change.
- The terminal result was `passed` with an unpublished pipeline head `cd5c90400c8e2781af59570cd42394e2b5c30162`; the required `no-mistakes axi sync --recover` returned custody cleanly. This is 2/5, leaving three loops unused. Ordinary push, #4059 body update, and #4059 exact-head CI are next.
