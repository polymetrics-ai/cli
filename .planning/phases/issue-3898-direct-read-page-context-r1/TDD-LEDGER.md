# TDD ledger — Issue #3898: direct-read page context

Every behaviour below started RED with retained failing evidence, then became
GREEN only after the same focused command passed.

**These tests assert RETURNED RECORD COUNTS against a known-larger fixture, never
exit status.** That is the whole point: the defect exited 0 while discarding
97.9% of a collection, so an exit-status assertion would have passed against it.

## Red evidence, verbatim

Captured from `go test ./internal/connectors/engine/ -run 'DirectRead...'`
before any implementation existed:

```
--- FAIL: TestOperationDirectReadReturnsEveryRecordForRootArray
    records = 30, want 120 (provider default page is 30 — a short result must never be reported as a complete one)
--- FAIL: TestOperationDirectReadReturnsEveryRecordForCursorEnvelope
    results = 30, want 120
--- FAIL: TestDirectReadReturnsEveryRecordForNestedCursor
    logs = 30, want 120
--- FAIL: TestOperationDirectReadReturnsEveryRecordForOffsetLimit
    results = 30, want 120
```

The fixture holds 120 records and serves 30 when the client sends no page-size
parameter — the exact shape that produced the live GitHub finding.

| ID | Guarantee | Red assertion | Status |
| --- | --- | --- | --- |
| P1 | Declared page, not provider default | A `page_number` read returns the provider's default 30 instead of the declared 100. | GREEN |
| P2 | Collection reachable by page number | Following `next_number` does not reach all 120 records. | GREEN |
| P3 | Cursor strategies hand back a token | A `cursor` read reports no `next_cursor`, or claims an addressable page number it does not have. | GREEN |
| P4 | Legacy executor parity | The non-operation `DirectRead` path truncates where the operation path does not. | GREEN |
| P5 | offset_limit is addressable | `--page 2` on an `offset_limit` strategy does not return the second window. | GREEN |
| P6 | Single object unaffected | A one-object read issues more than one request, or is reported as an incomplete collection. | GREEN |
| P7 | Undeclared paging is admitted | A bundle declaring no strategy reports `complete: true` it cannot prove. | GREEN |
| P8 | Refusal, never a quiet page one | Asking a cursor strategy for `--page 3` silently returns page one instead of erroring. | GREEN |
| P9 | Cursor onto a non-final page | Following a cursor onto a page that still has successors faults on a nil loop-guard map. | GREEN |

P9 exists because an earlier test passed while hiding a real panic: its followed
page happened to be the last one. Reverting the fix reproduces
`panic: assignment to entry in nil map` in `tokenPathCursor.Next`, which is the
retained red evidence for that row.

## Captain validation regression: limited reverse plan

The captain-required private-repository validation staged one row from the
three-row sample table. `pm reverse plan` succeeded, but `pm reverse run`
rejected the untouched source before any GitHub request. The local regression
below reproduces the same condition without credentials or a network write:

```
--- FAIL: TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice (1.43s)
    reverse_confirmation_test.go:192: PreviewReversePlan() error = reverse plan source rows or payload files changed before preview
FAIL
```

The red test plans `Limit: 1` from a two-record warehouse fixture, then
requires preview and run to stage exactly one record. Before the correction,
preview and execution instead read `RecordCount + 1`, hashing a second,
unapproved record as if it were drift. The pre-existing changed-row rejection
remains a separate green regression.

Green after changing both preview and run to read `max(1, RecordCount)`:

```sh
go test -timeout 20m ./internal/app -run '^(TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice|TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange)$' -count=1
# ok   polymetrics.ai/internal/app
```

## Live GitHub write-target investigation

The captain-authorized live `create_issue` dispatch returned
`Post "https://api.github.com/.../issues%22: EOF` before a private-repository
mutation. A local regression was added to assert the full generic reverse
workflow sends exactly `POST /repos/acme/widgets/issues`. It is GREEN on the
current implementation:

```sh
go test -timeout 20m ./internal/app -run '^TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint$' -count=1
# ok   polymetrics.ai/internal/app
```

This does **not** reproduce the live EOF, so it is recorded as an environment
or transport-path finding rather than papered over with a parallel client. No
production code was changed for that observation; a live retry is only safe
after independently confirming the private repository still has zero issues.

## Red/green commands

```sh
go test ./internal/connectors/engine/ -run 'DirectRead' -count=1
go test ./internal/connectors/commandrunner/ -count=1
go test ./internal/cli/ -run 'DirectRead|ConnectorCommand|Manual|Help|Limits' -count=1
```

No live provider call is made by any test in this ledger; every fixture is a
local `httptest` server with fabricated record identifiers.
