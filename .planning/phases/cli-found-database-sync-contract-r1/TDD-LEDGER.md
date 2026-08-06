# TDD ledger — Issue 3810: shared database sync contract

Manual GSD TDD execution; red output is retained in `traces/red-run.txt` before production code.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Closed modes | No shared type accepts exactly the seven names. | `Mode.Validate` accepts only `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`. |
| R2 | No capability without execution | A declared mode can appear executable with no native runtime/evidence. | Native admission requires a matching named executor and complete embedded-fixture evidence; the zero executor is false. |
| R3 | Versioned envelope | A stream checkpoint is only a scalar cursor. | Envelope round-trip has state version, source identity, mechanism, barrier, primary/tie-breaker, separate partitions, schema/protocol, dedupe identity/window, observed and committed times. |
| R4 | Opaque tokens | JSON/state handling can alter non-text token bytes. | Non-UTF-8 token bytes round-trip exactly and all public copies are independent. |
| R5 | Partition safety | Per-partition positions can be joined into one string or duplicate silently. | Partition records remain distinct, duplicate IDs are rejected, and resume checks preserve the complete list. |
| R6 | Explicit recovery | Invalid state can fall through to a new scan. | Each invalid checkpoint/retention/slot/token/generation/identity condition returns typed `RebootstrapRequiredError`; validation changes no state. |
| R7 | Commit order | A successful read or a partial destination result can call state persistence before sink acknowledgement. | Commit callback is never invoked without durable acknowledgement and runs only after an acknowledgement timestamp; `committed_at` is separate from `observed_at`; a `WriteResult` with failed records leaves state unchanged. |
| R8 | Delete semantics | A delete lacks stable identity/image/order rules and can request physical deletion in history mode. | Tombstone validation requires identity/key/image/order; history resolution exposes `_valid_from`, emits a `_valid_to`/`_is_current=false` validity-window close, and rejects physical target delete. |
| R9 | Native command boundary | A native command can smuggle REST surface or raw query text. | Native contract exposes only fixed protocol/executor/mode/evidence fields and rejects missing executor/evidence; no API-surface declaration is created. |
| R10 | Legacy persisted state | A state JSON `cursor` silently resumes under the new code. | Old JSON decodes to version-zero legacy state and `RunETL` receives a typed rebootstrap outcome before calling the source. |
| R11 | Existing legacy behavior | Existing local append/overwrite/incremental tests depend on `StreamState.Cursor`. | Tests retain observed behavior while inspecting the envelope's primary checkpoint instead. |
| R12 | Shared fixture corpus | Engine lanes could omit core change, replay, and handoff scenarios. | The immutable version-one corpus requires insert, update, both delete forms, truncate/invalidation, duplicate replay/dedupe, snapshot-to-stream handoff, recovery, and history-window cases. |

## Red command

```sh
go test ./internal/synccontract ./internal/app -run 'Test(Mode|Checkpoint|Resume|Commit|Tombstone|Native|Legacy)' -count=1
```

The initial run is expected to fail before the new package and state shape exist. The exact output
is retained, then the same focused command must pass after implementation.

## Green evidence

Passed after implementation:

```sh
go test ./internal/synccontract -count=1
go test ./internal/app -count=1
```

`internal/synccontract/contract_test.go` covers R1–R9. `internal/app/sync_state_test.go` covers
the version-zero legacy migration, no-read/no-clear rebootstrap behavior, post-ack state storage,
contract-mode refusal, and partial downstream result refusal. The pre-existing scalar-field
assertion in `internal/app/sync_modes_test.go` was deliberately changed to inspect the envelope's
primary position, retaining the same failed-run non-advancement assertion.

## Follow-up contract-audit red/green evidence

The research envelope lists a distinct `dedupe_window`. Before adding it, the focused assertion
failed to compile because `CheckpointEnvelope.DedupeWindow` and `HistoryValidFromColumn` did not
exist:

```sh
go test ./internal/synccontract -run 'TestCheckpointEnvelopePreservesOpaqueTokensAndPartitionState|TestTombstoneClosesHistoryWindowInsteadOfPhysicalDelete|TestConformanceFixturesAreVersionedAndDefensivelyCopied' -count=1
# ... DedupeWindow undefined ...
```

After adding the opaque start/end window, `_valid_from`, and the missing reusable scenarios, the
same package passed with `go test ./internal/synccontract -count=1`.
