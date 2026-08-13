# #4081 — TDD ledger

**Status:** Green implementation is locally verified. The remaining lifecycle
steps are final committed-head binary evidence, `verify-work`, code review,
no-mistakes, draft-PR checks, and CI; they are not silently claimed here.

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

## Carrier GREEN contract

The uncommitted Green slice following `168420e6b` adds only the accepted
carrier and its narrow supporting readback schema:

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

All commands above passed in the current worktree. A fresh committed-head
binary digest/size and the same faithful-server lifecycle must still be added
to `VERIFICATION.md` immediately after the Green commit; this ledger does not
substitute a dirty-worktree identity for that final evidence.

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
