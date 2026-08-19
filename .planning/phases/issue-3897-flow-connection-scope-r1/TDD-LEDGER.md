# #3897 TDD Ledger

**Status:** GREEN verified locally; correction 3 / 5 applied; delivery/checkpoint gates pending.
**Correction rounds:** 3 / 5

| Slice | Red evidence | Green evidence | Refactor / result |
|---|---|---|---|
| 1. Selected source rows | **RED 2026-08-11:** `go test -timeout 20m ./internal/cli -run '^TestFlowSourceConnectionSelectorsReadOnlyOwningRows$' -count=1` failed for both explicit selectors. DuckDB reported bare `records` absent and suggested only a qualified owner view. | **GREEN:** the same focused test passes: `connection: "acme"` returns only `acme-1` through a real Parquet/DuckDB read; `action_cfg.source_connection: "globex"` returns only `globex-1` through `QueryTableRequest.Connection`. | Replaced action SQL string construction with the existing table request; selector conversion is centralized in the flow CLI adapter. |
| 2. Omitted and root selectors | **RED:** the initial same-name materializations had no selectable bare source read, establishing that an implicit owner cannot be accepted. | **GREEN:** `TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed` verifies typed omissions with manifest-only remedies and root-only `_unattributed` query/action rows. | `warehouse.FindTable` remains the ownership authority; flow only decorates its typed error. |
| 3. Serialization/action boundary | **RED:** the initial manifest had no action source selector to parse, serialize, or preserve at the runner boundary. | **GREEN:** focused CLI and flow-engine tests JSON round-trip both fields, assert the selected request connection, and assert the local action runner receives `SourceConnection`. | No preview/digest exists in this flow path; #3994 owns later action lifecycle work. |
| 4. Public proof | **RED:** before implementation, the first focused flow test could not select either owner. | **GREEN:** a freshly built binary materialized same-named rows, ran a `connection: "acme"` flow query whose SQL fails if the wrong row is visible, and asserted returned `acme`/`globex` query rows. | Runtime help/manual, website docs, and three golden help surfaces document manifest syntax only; no flow CLI flag was added. |
| 5. Correction 1 — complete action source | **RED 2026-08-11:** `go test -v -timeout 20m ./internal/cli -run '^TestFlowActionSourceReadsAllSelectedConnectionRows$' -count=1` failed: each action runner received 100 rather than 101 selected `acme` rows, while `App.QueryTable(..., Limit: 0)` correctly remained capped at 100. | **GREEN 2026-08-11:** focused app/flow/CLI verification passed; successful and locally failed action attempts each received all 101 selected rows, and only the failed attempt lacked a success checkpoint. | `ActionSourceReadRequest` has no limit; `ReadActionSource` is used only by the flow action adapter, while public `QueryTable` retains its 100-row default. |
| 6. Correction 2 — fault-aware quoted DuckDB bindings | **RED 2026-08-11:** `go test -v -timeout 20m ./internal/app -run '^TestQuerySQLRefusesUnscopedHealthyAndUnreadableOwnerCollision$' -count=1` failed because unscoped `records` returned success after a competing owner record became unreadable. | **GREEN 2026-08-11:** focused app verification passes the hidden-owner refusal, selected and `_unattributed` reads, quoted `1orders`/`orders-2026`/`orders.2026` reads, typed omitted ambiguity, aggregate, and cancellation paths. | Bare views use `warehouse.FindTable`; the pinned parsed-table replacement callback preserves the first original typed lookup error, and DuckDB identifiers quote embedded double quotes. |
| 7. Correction 3 — one resolver per query | **RED 2026-08-11:** `TestQuerySQLReusesOneWarehouseResolverPerQuery` failed with three resolver builds for two selected tables, proving repeated full inventory scans. | **GREEN 2026-08-11:** the focused race matrix passes one-snapshot multi-table and replacement-scan tests together with all R1/R3/R4, aggregate, cancellation, and warehouse fault coverage. | `warehouse.TableResolver` captures immutable tables/faults once, indexes names, and supplies all query view and callback resolution; `FindTable` delegates to the same rules. |

## Red: required first executable test

Create `acme` and `globex` materializations through normal local ETL with the
same `records` table name. The test must drive the flow query and action source
read boundaries with an explicit declared selector and assert the actual rows
from the Parquet/DuckDB path. It is RED until the current selector drop is
removed. No provider action, approval token, or external network call is part
of the test.

### Recorded RED output

`TestFlowSourceConnectionSelectorsReadOnlyOwningRows` creates two connection
records plus two `warehouse.Location` ownership records and writes each
`records.parquet` table through `warehouse.WriteTable`. The query and action
source reads both failed at the real DuckDB boundary:

```text
flow: step failed: step <step>: execute query: Catalog Error:
Table with name records does not exist!
Did you mean "records__conn_<owner-id>"?
```

This is the target defect: the declared `connection: "acme"` and JSON
`action_cfg.source_connection: "globex"` were not carried into the source
read. The full non-secret result is in `traces/red-flow-source-selectors.txt`.

## Green safety constraints

- `warehouse.FindTable` remains the owner-selection authority.
- `_unattributed` is forwarded unchanged and never becomes a real connection.
- SQL receives no user-built identifier interpolation.
- Omitted selection never falls back to an arbitrary owner.
- An action remains a local stub boundary; #3994 owns dispatch.

## Executed GREEN commands

- `go test -timeout 20m ./internal/cli -run '^(TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed)$' -count=1`
- `go test -timeout 20m ./internal/app -run '^(TestQuerySQLScopesConnectionOwnedAndUnattributedViews|TestQuerySQLAmbiguityNamesNoSelectorItCannotAccept|TestQuerySQLHonorsCanceledContext)$' -count=1`
- `go test -timeout 20m ./internal/app -count=1`
- `go test -timeout 20m ./internal/flow -count=1`
- `go test -timeout 20m ./internal/cli -count=1`
- `go test -race -timeout 20m ./internal/app ./internal/flow -count=1`

All listed commands passed. The first RED output remains in
`traces/red-flow-source-selectors.txt`; the non-secret binary proof is in
`traces/binary-flow-proof.txt`.

## Correction 1 — #4032 complete action source

The correction RED fixture replaced the `acme` table with 101 structurally
owned rows while retaining the `globex` owner. It uses the production flow
adapter and a local capture runner, so it proves full selected-row delivery,
typed owner selection, and no provider mutation. The pre-fix action boundary
called public `QueryTable(..., Limit: 0)`, which intentionally applies its
100-row CLI default; both the successful and locally failed action subtests
therefore captured only 100 rows. The public cap is asserted separately in the
same fixture.

**GREEN:** `go test -v -timeout 20m ./internal/app ./internal/flow
./internal/cli -run
'^(TestEnginePassesManifestSourceConnectionSelectors|TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed|TestFlowActionSourceReadsAllSelectedConnectionRows)$'
-count=1` passed. It compiles the new app/flow contracts, preserves the
existing selection/ambiguity/root tests, and proves both 101-row action paths.

## Correction 2 — #4037 DuckDB warehouse binding

**RED:** `TestQuerySQLRefusesUnscopedHealthyAndUnreadableOwnerCollision`
failed with a nil error because the healthy `acme/records` subset established a
bare view after the competing `globex/records` owner record became unreadable.

**GREEN:** `go test -v -timeout 20m ./internal/app -run
'^(TestQuerySQLScopesConnectionOwnedAndUnattributedViews|TestQuerySQLAmbiguityNamesNoSelectorItCannotAccept|TestQuerySQLRefusesUnscopedHealthyAndUnreadableOwnerCollision|TestQuerySQLBindsQuotedConnectionScopedWarehouseNames|TestWarehouseQueryIdentifierQuotingAndIdentityValidation|TestQuerySQLAggregatesOverParquetTables|TestQuerySQLHonorsCanceledContext)$'
-count=1` passed. `FindTable` now decides all bare views, while the pinned
`RegisterReplacementScan` callback receives parsed logical table names and
preserves the first `FaultError` or `AmbiguousTableError` with `%w`. The local
fixtures write only warehouse Parquet data and do not mutate a provider.

## Correction 3 — #4040 resolver reuse

**RED:** `go test -json -timeout 20m ./internal/app -run
'^TestQuerySQLReusesOneWarehouseResolverPerQuery$' -count=1` failed with
`warehouse resolver builds = 3, want one inventory snapshot per query`.

**GREEN:** the focused race matrix passed with one resolver build for a
two-table query and one resolver build when an omitted duplicate enters the
replacement scan. The resolver's direct warehouse test retains undecided
faults, explicit selection, unaffected healthy tables, and immutable snapshot
visibility; the existing action source, scope, ambiguity, quoted-name,
aggregate, and cancellation tests remain green.
