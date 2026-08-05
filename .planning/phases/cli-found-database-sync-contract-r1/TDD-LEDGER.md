# TDD ledger — Issue 3810: shared database sync contract

Manual GSD TDD execution; red output is retained in `traces/red-run.txt` before production code.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Closed modes | No shared type accepts exactly the seven names. | `Mode.Validate` accepts only `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`. |
| R2 | No capability without execution | A declared mode can appear executable with no native runtime/evidence. | Native admission requires a matching named executor and complete embedded-fixture evidence; the zero executor is false. |
| R3 | Versioned envelope | A stream checkpoint is only a scalar cursor. | Envelope round-trip has state version, source identity, mechanism, barrier, primary/tie-breaker, separate partitions, schema/protocol, dedupe, observed and committed times. |
| R4 | Opaque tokens | JSON/state handling can alter non-text token bytes. | Non-UTF-8 token bytes round-trip exactly and all public copies are independent. |
| R5 | Partition safety | Per-partition positions can be joined into one string or duplicate silently. | Partition records remain distinct, duplicate IDs are rejected, and resume checks preserve the complete list. |
| R6 | Explicit recovery | Invalid state can fall through to a new scan. | Each invalid checkpoint/retention/slot/token/generation/identity condition returns typed `RebootstrapRequiredError`; validation changes no state. |
| R7 | Commit order | A successful read can call state persistence before sink acknowledgement. | Commit callback is never invoked without durable acknowledgement and runs only after an acknowledgement timestamp; `committed_at` is separate from `observed_at`. |
| R8 | Delete semantics | A delete lacks stable identity/image/order rules and can request physical deletion in history mode. | Tombstone validation requires identity/key/image/order; history resolution emits a validity-window close and rejects physical target delete. |
| R9 | Native command boundary | A native command can smuggle REST surface or raw query text. | Native contract exposes only fixed protocol/executor/mode/evidence fields and rejects missing executor/evidence; no API-surface declaration is created. |
| R10 | Legacy persisted state | A state JSON `cursor` silently resumes under the new code. | Old JSON decodes to version-zero legacy state and `RunETL` receives a typed rebootstrap outcome before calling the source. |
| R11 | Existing legacy behavior | Existing local append/overwrite/incremental tests depend on `StreamState.Cursor`. | Tests retain observed behavior while inspecting the envelope's primary checkpoint instead. |

## Red command

```sh
go test ./internal/synccontract ./internal/app -run 'Test(Mode|Checkpoint|Resume|Commit|Tombstone|Native|Legacy)' -count=1
```

The initial run is expected to fail before the new package and state shape exist. The exact output
is retained, then the same focused command must pass after implementation.
