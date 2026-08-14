# TDD ledger — Issue 3973: transactional database write sessions

Manual inline GSD TDD execution. Red and green command output is retained in
`traces/` before and after production changes.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Approval admission | A target/schema/mode/key/count/effects mismatch can reach a driver batch. | Every mismatch is refused with zero fake driver begin/batch/commit/rollback calls. |
| R2 | One-shot approval | Replaying approval can perform a second mutation. | First accepted execution consumes approval before its first batch; a replay has zero new session/batch calls. |
| R3 | Pinned bounded execution | Per-record auto-commit or multiple sessions can make a batch run succeed. | Six rows at a two-row limit produce one session and batches `[2,2,2]`; legacy write count remains zero. |
| R4 | Whole-session rollback | Batch failure/cancellation can leave partial work acknowledged. | A later batch failure and cancellation each cause one rollback, no commit/receipt/checkpoint authority. |
| R5 | Commit certainty | An indeterminate commit can be retried or reported as rolled back. | Outcome is `unknown`; commit/batch/rollback counters do not change again and no receipt/checkpoint authority exists. |
| R6 | Mode safety | `full_overwrite` can publish non-atomically or canonical modes can use the wrong strategy. | Non-atomic overwrite is refused before begin; successful modes record only canonical strategies through the one pinned session. |
| R7 | Durable checkpoint gate | A checkpoint can advance before durable target commit evidence. | Only a confirmed durable receipt produces acknowledgement authority; all unsuccessful states produce none. |

## Red command

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1
```

**Red:** The base has no `DatabaseWriteExecutor`/`WriteSession` contract, so
the newly added behavioural test must fail to compile or execute. The exact
result is retained in `traces/write-session-red.txt`.

## Green commands

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1
go test -race -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1
go test -timeout 20m ./internal/connectors/database/... ./internal/app/...
```

**Green:** The fake exposes each observable mutation and the test asserts its
session count, bounded batches, approval ordering, outcome, receipt, and
checkpoint eligibility rather than merely checking returned errors.

## Recorded result

- **Red:** `traces/write-session-red.txt` records the missing-session-contract
  build failure before the production type existed.
- **Green:** `traces/write-session-green.txt` records the focused normal and
  race runs after the implementation. The suite observes zero driver/session/
  ledger calls for refused plans and every required lifecycle counter for
  successful, rolled-back, and unknown outcomes.
