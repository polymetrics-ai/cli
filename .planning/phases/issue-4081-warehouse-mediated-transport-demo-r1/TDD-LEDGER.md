# #4081 — TDD ledger

**Status:** Green implementation, the final connector-boundary correction, and
their local gates are verified. The remaining lifecycle steps are no-mistakes,
draft-PR checks, CI, and automated-review disposition; they are not silently
claimed here.

**Base admission:** `docs/4015-connector-release-certification` at
`e7d2b2963fc1dd164f63b31fccb8a3bab8084bec`, with #4019's squash-admitted
#4077 Transport content proven by exact blobs rather than ancestry.

## Committed RED → GREEN sequence

| Commit | TDD checkpoint | Evidence / preservation |
| --- | --- | --- |
| `1adc808e4` | Plan | Issue/GSD context, plan, checkpoint, ledger, and verification artifacts landed before production work. |
| `7d0430f7b` | RED — construction and durable handoff | Focused app/synctransport tests exposed missing production construction, durable reopen, tamper rejection, and receipt-before-CAS contracts. |
| `8fdff30b6` | GREEN — shared contract | Added `WarehouseReceipt`, connection identity, reopen and independent read-back contracts without an app provider composition. |
| `3fb1b939e` | GREEN — durable stage | Added owner-scoped WAL → DuckDB/Parquet → immutable-manifest receipt/reopen and made durable-reopen tests pass. |
| `cbbe4042c` | GREEN — provisional composition | Added GitHub source/destination composition, intentionally left a direct-write approval gap for the next RED. |
| `b5bbf12e0` | RED — approval gate | Proved that an unbound direct transport write was not approval-bound. |
| `d6a81e93d` | GREEN — pre-run binding | Added the closed plan → preview → one-time grant lifecycle and workset/receipt-derived forward binding; Transport itself does not mint approval. |
| `589fe8c4e` | RED evidence extension | Proved fresh `App.Open` restart isolation, mutable returned-value isolation, non-nil empty slices, and physical manifest/WAL/Parquet corruption rejection. |
| `b5ccedefd` | Compatibility refactor | Kept legacy GitHub JSON ETL on its old route; only the exact persisted demo connection selects Transport. |
| `4e1f9e489` | GREEN — correction loop **3/5** | Authenticates cleanup's forward plan seal, plan hash, runtime identity, and rederived binding before minting typed inverse authority; tamper and 404/replay tests pass. |
| `f9bd77751` | Superseded carrier RED | Historical one-shot `pm demo` proposal; retained as history but deliberately not made Green after the carrier decision rejected that surface. |
| `168420e6b` | Accepted carrier RED | Replaced the rejected one-shot test with separate closed `pm etl transport` plan/preview/cleanup subprocesses and an ordinary approval-carrying `pm etl run`. It failed because the plan command exited `2` (`invalid usage`) and an approval-bearing legacy ETL still reached source `Read`. |

## Carrier GREEN contract — `6220144db`

The Green slice following `168420e6b` adds only the accepted carrier and its
narrow supporting readback schema:

- `pm etl transport github-issue-label` creates/preview plans and derives only
  a fixed cleanup inverse; it exposes no provider URL, action, record, repo,
  issue, label, base URL, or credential input.
- `pm etl run` accepts a closed transport tuple only when it has exact
  `--connection`, `--stream issues`, `--batch-size 1`, `--approval-plan`, bare
  `--approval-token-stdin`, and `--confirm destructive`. The token is read as
  exactly one bounded stdin line after the tuple has been validated. A raw token
  flag, token-like value, repeated selector, caller-selected provider field,
  partial tuple, runtime flag, wrong stream/batch size, or wrong confirmation
  fails closed.
- `App.RunETL` rejects non-empty destination approval material before legacy
  source I/O unless the resolved persisted connection is the exact closed
  GitHub issue-label transport route.
- The declarative GitHub `issues` schema now declares label names so an
  independent engine `Read` can actually observe the typed mutation. This is
  not a generic writer or a new connector framework.
- The faithful server proves the source → WAL/DuckDB/Parquet stage → reopened
  singleton → typed POST → independent GET → acknowledgement → checkpoint
  order; independently planned cleanup proves DELETE `204`, separately planned
  idempotent DELETE `404`, replay rejection, and no remaining label.

### GREEN commands observed before the implementation commit

```text
go test -count=1 -timeout 20m -run '^(TestETLTransportBareAndLeafHelpAreContextual|TestETLRunTransportApprovalRejectsUnsafeOrIncompleteCarriage|TestETLRunTransportApprovalReadsExactlyOneEphemeralStdinLine|TestETLTransportSafeOutputOmitsApprovalAndDestinationInternals|TestPMBinaryExecutesGitHubWarehouseTransportLifecycle|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -v ./internal/cli
go test -count=1 -timeout 20m -run '^TestRunETLRejectsDestinationApprovalOutsideClosedTransportRoute$' ./internal/app
go test -count=1 -timeout 20m ./internal/app
go test -race -count=1 -timeout 20m -run '^TestGitHubIssueLabel' -v ./internal/app
go test -count=1 -timeout 20m -run '^TestGithubPullRequestsETLSupportsAllSyncModes$' -v ./internal/app
go test -count=1 -timeout 20m ./internal/synctransport
go test -count=1 -timeout 20m ./internal/connectors/engine
go test -count=1 -timeout 20m ./internal/connectors/commandrunner
go test -count=1 -timeout 20m ./internal/cli
go vet ./internal/app ./internal/cli ./internal/synctransport ./internal/connectors/engine ./internal/connectors/commandrunner
go run ./cmd/connectorgen validate internal/connectors/defs/github
go run ./cmd/connectorgen surface-sync --check
git diff --check
```

All commands above passed in the current worktree. The fresh committed-head
binary digest/size and faithful-server lifecycle are recorded in
`VERIFICATION.md`; the committed-tree repetition passed at `6220144db`.

## Correction loop 5/5 — definition-owned boundary selection

The final allowed correction was a real #4081 violation, not baseline debt.
Before mutation, the exact `go run ./cmd/connectorgen boundary . --json` rule
was compared in isolated worktrees:

```text
admitted parent e7d2b2963fc1dd164f63b31fccb8a3bab8084bec: exit 0, outcome clean, findings 0
pre-correction #4081 head 6220144db21e4170bb400d3e5aefd65d04b4111e: exit 1, outcome policy_violations, findings 378
```

The introduced findings were confined to the #4081 composition/carrier paths:
`internal/app/app.go`, `internal/app/github_transport_approval.go`,
`internal/app/github_warehouse_transport.go`, `internal/app/transport_dispatch.go`,
`internal/cli/cli.go`, and `internal/cli/etl_transport.go`. No exception entry
was added or changed.

GREEN moves provider selection to the existing declarative definition: App
finds the unique engine connector with the closed `issues`, `add_issue_labels`,
and `remove_issue_label` contract, wraps the concrete engine connector only to
attach the fixed Transport descriptor, and passes the selected identity through
the closed carrier. Shared App/CLI code contains no provider endpoint/action
policy; the definition still owns it. The carrier remains exactly
`pm etl transport github-issue-label`, but its visible identity is derived from
the selected definition rather than hard-coded into shared production code.

The first neutral selector incorrectly treated a one-sided descriptor as a
Transport pair; RED from the unchanged legacy five-mode suite exposed that
coupling. The narrowed dispatch now returns to ordinary JSON ETL unless both
ends are the complete closed pair. This is the final correction in the 5/5
budget; no exception or unrelated compatibility rewrite was used.

### Final correction GREEN evidence

```text
make connector-boundary
  outcome: clean; findings: 0
go test -count=1 -timeout 20m -run '^TestGithubPullRequestsETLSupportsAllSyncModes$' -v ./internal/app
  PASS: full_refresh_append_duplicates, full_refresh_overwrite_replaces_final,
  full_refresh_overwrite_deduped_keeps_latest_duplicate,
  incremental_append_filters_older_cursor_and_appends_inclusive_cursor,
  incremental_append_deduped_materializes_latest_PR_rows
go test -race -count=1 -timeout 20m -run '^TestIssueLabelTransport' -v ./internal/app
  PASS
go test -count=1 -timeout 20m -run '^(TestETLTransportBareAndLeafHelpAreContextual|TestGoldenTranscripts)$' -v ./internal/cli
  PASS
go test -count=1 -timeout 20m ./internal/app ./internal/cli ./internal/synctransport ./internal/connectors/engine ./internal/connectors/commandrunner
  PASS
go test -count=1 -timeout 20m -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$' -v ./internal/cli
  PASS; sha256=3af708eb697e566edd16480bff9f9bf345ea90ab8c5b25e3ee4a0689ae5e827d; size=148530114
go mod tidy -diff; go vet ./internal/app ./internal/cli ./internal/synctransport ./internal/connectors/engine ./internal/connectors/commandrunner; make lint
  PASS
```

## Refactor and safety preservation

- No obsolete receipt-time planner, synthetic workset, arbitrary table
  rematerializer, generic HTTP writer, or generic inverse remains; the exact
  removal check was run against implementation, docs, and phase evidence with
  no matches.

- The rejected one-shot file and temporary App demo helper were removed only
  while uncommitted; no pushed history was reset, cleaned, or force-pushed.
- TOCTOU locking, strict hash syntax validation, and `O_EXCL` collision
  reservation remain explicitly parked for the broader production Transport
  MVP, as do PostgreSQL legs, schedules/flows/CDC, auth/rate coordination,
  exhaustive certification, generator promotion, and release promotion.
