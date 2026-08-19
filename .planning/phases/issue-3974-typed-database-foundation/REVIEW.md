# Code review — Issue #3974: typed database connector foundation

## Captain amendment status

The original review predates the captain's warehouse-mediation ruling. The
amended `warehouse` and `database` seam has now received a refreshed scope,
security, and type-boundary review after its local gates passed.

## Method

`scripts/gsd prompt code-review 3974 --auto` was generated and reviewed. Its
normal runtime requires a numbered phase with a generated isolated reviewer;
this issue is a non-numbered foundation slice and the task contract does not
provide that role. The repository-approved manual inline fallback was used.

Reviewed the changed foundation files for:

- scope leakage into SQL execution, target DDL/provisioning, write sessions,
  receipts/checkpoints, CDC, or capability promotion;
- closed schema/decoder behavior, secret-safe errors, defensive projections,
  cancellation, finite resource limits, and typed identity boundaries;
- cross-package admission semantics so source execution is not implied by a
  database target admission; and
- PostgreSQL bundle/reference seam compatibility with the existing engine.
- warehouse mediation shape: a connector can name only one of its own legs,
  and a single native descriptor cannot stand in for both directions.

## Findings and disposition

| ID | Severity | Finding | Disposition |
| --- | --- | --- | --- |
| R1 | warning | Direct lint initially found an ineffective empty-mode mutation test, an avoidable struct conversion, and an unused target validator. | Fixed before final validation; direct lint and focused tests pass after the refinement. |
| R2 | pass | No executable SQL, generic query/REST-write path, target DDL, write session, receipt/acknowledgement, CDC implementation, or mode enum was introduced. | Confirmed by source review and the final scope scan. |
| R3 | pass | The PostgreSQL declaration remains non-executing and metadata still reports `write=false`, `cdc=false`. | Confirmed by `database_driver_test.go` and `database_definition_test.go`. |
| R4 | pass | Admission remains fail-closed: declaration, registry identity, protocol/API match, and shared evidence must all agree; a stored admission without `RunNativeSync` cannot source-dispatch. | Confirmed by focused database and synccontract tests. |
| R5 | pass | The original singular driver admission shape would have encouraged one descriptor to claim both source-to-warehouse and warehouse-to-target work. | Replaced with separate `DatabaseNativeAdmissions`; the registry requires an exact per-leg descriptor, and an inbound-only fixture is refused for an outbound command. |
| R6 | pass | Layer one must not acquire PostgreSQL/MySQL mechanics. | `go list -deps` confirms `internal/warehouse` and `internal/connectors/database` import neither native database driver; the MySQL type-level proof changes no shared package. |

## Final verdict

No unresolved critical, warning, security, or scope findings remain in the F1
slice or captain amendment. The isolated CLI regression suite, focused/race
tests, and every required broad static/build gate passed after the resolved
per-leg-admission refinement.
