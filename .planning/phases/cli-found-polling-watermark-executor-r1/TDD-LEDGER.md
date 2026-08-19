# TDD LEDGER — polling-watermark changefeed executor

Manual GSD execution: each production behavior begins with a recorded failing
test. No provider, credential, database, or wall-clock sleep is used.

| ID | Behavior | Required red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | Test bundle is CDC-capable only through its matching executor | an implemented `polling_watermark` declaration alone remains non-capable | the existing real projection reports CDC only with a full matching executor |
| R2 | Complete checkpoint contract is mandatory | a declaration/executor missing `kind`, `keys`, `commit_after`, or `on_invalid` is accepted or promoted | each omission rejects/does not promote |
| R3 | Ties replay safely | page-edge records sharing a watermark are skipped or `at_least_once` is not enforced | inclusive boundary rereads edge records and delivers duplicates truthfully |
| R4 | Timestamp safety lag includes late arrivals | checkpoint-derived lower bound omits an earlier timestamp | committed state produces a lagged lower bound; an initial read leaves its snapshot boundary to the source |
| R5 | Delete observability is honest | a hard-delete-only declaration claims a tombstone or a soft-delete marker is ignored | `not_available` remains non-tombstone; declared marker yields tombstone |
| R6 | Checkpoint follows durable destination acknowledgement | destination failure or persistence failure advances the checkpoint | no advance on failure; post-accept persistence failure reruns the same window |
| R7 | Work and cancellation are bounded | executor exceeds page/request limits or ignores cancelled context | bounded calls stop at every limit; cancellation returns without a follow-up fetch |

## Run log

### R1–R7 — red confirmed

Before production changes, the focused semantic suite was added with a
test-only implemented declaration. It names the absent shared executor,
page-source port, and checkpoint committer first so the
eventual implementation cannot silently fall back to the old scalar cursor
path.

```text
$ go test ./internal/connectors/engine -run '^TestPollingWatermark' -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
polling_watermark_test.go:20:13: undefined: PollingWatermarkPage
polling_watermark_test.go:21:13: undefined: PollingWatermarkPageRequest
polling_watermark_test.go:74:60: undefined: PollingWatermarkPageSource
polling_watermark_test.go:74:118: undefined: PollingWatermarkConnector
FAIL    polymetrics.ai/internal/connectors/engine [build failed]
```

No production code had been edited at this point. The new fixture also carries
the intentionally unsupported `polling_watermark` JSON fields, so after the
missing executor API is supplied the loader test will remain red until the
declaration schema and semantic validation are extended.

### R1–R7 — green transition

The shared executor now consumes only the declaration, a closed page source,
the destination-acknowledging event callback, and the narrow
checkpoint committer. Its tests prove all required paths without a provider,
credential, database, or sleep:

```text
$ go test ./internal/connectors/engine -run '^TestPollingWatermark' -count=1
ok      polymetrics.ai/internal/connectors/engine

$ go test ./internal/connectors/engine -count=1
ok      polymetrics.ai/internal/connectors/engine

$ go test ./internal/connectors -count=1
ok      polymetrics.ai/internal/connectors
```

- R1 proves the test-only bundle has no CDC capability as a plain declarative
  connector and gains it only from the matching shared executor; definition,
  registry list/catalog, and manifest use the existing single projection gate.
- R2 loads a declaration with every checkpoint field omitted in turn and proves
  rejection through the real loader schema/semantic path.
- R3 records `a,b,b,c` across the inclusive timestamp boundary, requires the
  tuple on the second source request, and advertises `at_least_once`.
- R4 uses a committed `09:59` timestamp to request `09:57` for its two-minute
  safety lag, while an initial read leaves its snapshot boundary to the source.
- R5 proves both soft-delete and fixed declared deletion-endpoint tombstones;
  a hard-delete-only descriptor claiming tombstones is rejected.
- R6 proves no commit on destination rejection and a checkpoint persistence
  failure replays its already accepted page on the next run.
- R7 proves the declaration's page size, request budget, maximum pages, and
  cancellation checks are honoured.
