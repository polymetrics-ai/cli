# Code review — Issue #3974: typed database connector foundation

## Captain amendment status

The original review predates the captain's warehouse-mediation ruling. This
review is reopened for the amended `warehouse` and `database` seam; its final
verdict is pending refreshed scope, security, and type-boundary review after
the amended gates complete.

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

## Findings and disposition

| ID | Severity | Finding | Disposition |
| --- | --- | --- | --- |
| R1 | warning | Direct lint initially found an ineffective empty-mode mutation test, an avoidable struct conversion, and an unused target validator. | Fixed before final validation; direct lint and focused tests pass after the refinement. |
| R2 | pass | No executable SQL, generic query/REST-write path, target DDL, write session, receipt/acknowledgement, CDC implementation, or mode enum was introduced. | Confirmed by source review and the final scope scan. |
| R3 | pass | The PostgreSQL declaration remains non-executing and metadata still reports `write=false`, `cdc=false`. | Confirmed by `database_driver_test.go` and `database_definition_test.go`. |
| R4 | pass | Admission remains fail-closed: declaration, registry identity, protocol/API match, and shared evidence must all agree; a stored admission without `RunNativeSync` cannot source-dispatch. | Confirmed by focused database and synccontract tests. |

## Original-slice verdict

No unresolved critical, warning, security, or scope findings remained in the
original F1 slice. The isolated CLI regression suite and every required broad
static/build gate passed after the resolved hygiene refinements. The amendment
must receive its own final verdict before this file is handed off.
