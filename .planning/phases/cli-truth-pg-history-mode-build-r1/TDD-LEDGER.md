# TDD ledger — PostgreSQL history-mode truthfulness repair

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Outer descriptor admits the claimed mode | The production-composed PostgreSQL preflight returns `source transport does not support sync mode "incremental_dedupe_history"`. | The same preflight returns the registered native PostgreSQL source/destination and descriptor-owned `dedupe_history` strategy. |
| R2 | Full CLI path seals the declared PostgreSQL history route | After outer admission, a freshly built `pm` reaches the target adapter but `etl run` fails with the typed history-route mismatch; after route forwarding, it exposes the missing history primary-key forwarding as `database write plan is invalid`. | The native source/destination adapters carry typed PostgreSQL definitions, the destination retains the declared key and source-owned candidate position, and the built binary completes the run. |
| R3 | Updates preserve history | Before repair the run cannot reach the existing close/insert writer. | A larger-cursor source update produces exactly one closed former row and one open current row with adjacent validity timestamps. |
| R4 | Replay is idempotent | Before repair no valid history snapshot exists to replay. | A later approved replay leaves the target's history rows byte-for-value equivalent; no duplicate or reopened row exists. |
| R5 | Forbidden routes remain safely closed | Existing fake source/destination route cells are rejected before session/ledger mutation. | Existing typed-error test stays green unchanged; it proves this repair does not expand other connector routes. |

## Red command

```sh
go test -timeout 20m -count=1 ./internal/app -run 'TestOpen(PostgresHistoryModeResolvesRegisteredExecutors|PostgresTransportDeclarationsAreExactModeIntersection)$'
```

The first red output is retained under `traces/` before descriptor edits.

**Red observed:** `TestOpenPostgresHistoryModeResolvesRegisteredExecutors`
failed with `source transport does not support sync mode
"incremental_dedupe_history"`; the exact-mode-intersection test independently
reported the missing mode in both outer lists. See
`traces/red-outer-history-admission.txt`.

**Sequencing note:** the existing audited binary reproduction is R2's red
evidence. The new tagged binary proof was added after R1's descriptor green
slice because it is a costly real-container test; it supplements, rather than
replaces, the direct pre-change refusal recorded above.

**Red observed — production seal:** after the descriptor green slice, the new
live built-binary test reached the managed target but failed before the first
target write with `database history route source and destination do not match
the declared managed-target driver`. The outer destination supplied neither
leg of the typed history route. After the route was forwarded, the same binary
test failed `database write plan is invalid`, exposing that its legacy key
selection omitted history mode. The green slice makes live and fixture runners
provide their immutable loaded definition, seals the outer route before adapter
I/O, retains history primary keys, and uses the durable source page tuple for
the existing conditional fence without adding a caller-selected driver
identity.

## Green commands

**Green observed:** `TestManagedTargetHistoryRouteUsesTypedPostgresDefinitions`,
`TestManagedTargetHistoryWriteUsesDeclaredPrimaryKey`, and
`TestManagedTargetHistoryWriteInputUsesSourceOwnedCandidatePosition` pass. The
tagged live binary test passes in 30.70 seconds and independently observes
initial, superseded, and replay-safe target history. See
`traces/green-binary-history-live.txt`.
