# PostgreSQL parity tree audit — captain addendum

- **Audited:** 2026-08-10
- **Live source:** GitHub REST read-back of #3972 and all eleven children after main advanced to
  `4df0b0416`.
- **Authority:** `data/captain.md`, the three accepted database/CDC reports, and
  `internal/synccontract/mode.go`.

## Finding

The original tree had the typed database, PostgreSQL read/write, CDC, workset, and certification
pieces, but it did not assign a pre-certification owner for the full warehouse-mediated flow/mode
matrix. Its parent and several child bodies also abbreviated routes as `API→PostgreSQL` or
`PostgreSQL→API`, which could be misread as a direct hop.

`#3974` is **not wrong**. It is the active typed-framework foundation and deliberately does not own
flow routing, database DDL, write sessions, or PostgreSQL CDC. The correction therefore leaves its
body unchanged and places warehouse-flow integration after the existing driver and CDC work.

## Required correction

Create a twelfth child from `POSTGRES-P6-ISSUE-BODY.md`, then revise the parent and final
certification dependency graph so the new matrix issue gates #3978. The graph change is deliberate:
the original graph let certification begin after P1–P5 even though no issue proved every source and
target side against the warehouse or classified every one of the seven modes.

The resulting execution order adds one integration wave between P1–P5 and final certification; it
does not gate active Wave A #3974 or change its scope.

## REST mutation/read-back result

- Created [#3987](https://github.com/polymetrics-ai/cli/issues/3987), `Postgres Parity: prove the
  warehouse-mediated flow and mode matrix`, and attached it under #3972. REST read-back reports
  **12 of 12** children.
- Rewrote [#3972](https://github.com/polymetrics-ai/cli/issues/3972) to name the four warehouse
  routes, all seven mode outcomes, the extra gate, and the revised execution waves.
- Rewrote [#3978](https://github.com/polymetrics-ai/cli/issues/3978) so it depends on #3987 and
  cannot certify a direct route or an unclassified mode.
- Added binding warehouse-route amendments to [#3981](https://github.com/polymetrics-ai/cli/issues/3981),
  [#3973](https://github.com/polymetrics-ai/cli/issues/3973),
  [#3975](https://github.com/polymetrics-ai/cli/issues/3975),
  [#3980](https://github.com/polymetrics-ai/cli/issues/3980),
  [#3976](https://github.com/polymetrics-ai/cli/issues/3976),
  [#3982](https://github.com/polymetrics-ai/cli/issues/3982),
  [#3977](https://github.com/polymetrics-ai/cli/issues/3977),
  [#3979](https://github.com/polymetrics-ai/cli/issues/3979), and
  [#3983](https://github.com/polymetrics-ai/cli/issues/3983). The comments preserve the original
  rationale while making the correction durable and visible.
- Deliberately made **no change** to active [#3974](https://github.com/polymetrics-ai/cli/issues/3974).
  Its source/target typed framework is correct; #3987 owns the later flow integration.

No PostgreSQL or GitHub parity implementation branch was checked out or edited. All external reads
and mutations used `gh-axi` REST routes; no GraphQL request was made.

## Per-issue disposition

| Issue | Disposition | Reason |
| --- | --- | --- |
| #3972 | Edit | Make all four warehouse paths, all seven mode classifications, and the new dependency explicit. |
| #3974 | No edit | Active Wave A foundation is correctly scoped; no direct-route contract belongs there. |
| #3981 | Comment amendment | Managed target must consume a sealed warehouse workset, not a source connector. |
| #3973 | Comment amendment | Write session input is warehouse-only; no direct ETL bypass. |
| #3975 | Comment amendment | CDC receipt/materialization ends in connection-owned warehouse state before acknowledgement. |
| #3980 | Comment amendment | Existing warehouse workset scope is retained and made the only outbound producer. |
| #3976 | Comment amendment | PostgreSQL reads must materialize into the warehouse before any target leg. |
| #3982 | Comment amendment | API read → warehouse → PostgreSQL write replaces ambiguous direct wording; non-supported modes fail closed. |
| #3977 | Comment amendment | CDC source delivery is to the warehouse transaction contract, never directly to a target. |
| #3979 | Comment amendment | Bootstrap's WAL/Parquet path is named as the connection-owned warehouse and no direct drain is allowed. |
| #3983 | Comment amendment | Derived PostgreSQL delivery is PostgreSQL → warehouse → PostgreSQL, not source Parquet direct to target. |
| #3978 | Edit | Final certification must consume the new matrix and prove all four routes plus all seven mode outcomes. |
| New P6 | Create and attach | Own the missing flow/mode conformance before #3978. |

## Mode truth used by the correction

| Mode | Correct PostgreSQL parity treatment |
| --- | --- |
| `full_overwrite` | Supported target mode after typed write/ownership/session foundations; warehouse workset required. |
| `full_append` | Supported target mode after the same foundations; warehouse workset required. |
| `incremental_append` | Supported target mode after the same foundations; warehouse workset required. |
| `incremental_upsert` | Supported target mode after the same foundations; key required; derived CDC worksets use it first. |
| `incremental_dedupe` | Supported target mode after the same foundations; deterministic winner required. |
| `incremental_dedupe_history` | Recognized vocabulary but explicitly non-executable in phase one; reject with a typed reason. |
| `change_capture` | PostgreSQL 14+ source-only CDC through streamed `pgoutput` v2, bounded stage, warehouse receipt before LSN acknowledgement; never cursor fallback or target mode. |
