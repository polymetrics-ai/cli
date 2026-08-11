# #4067 execution record

**Status:** local implementation, focused expansion, heavy validation, and manual review completed — fresh no-mistakes, existing-PR delivery, and exact-head CI remain pending.

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
