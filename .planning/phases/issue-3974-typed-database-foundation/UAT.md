# Verification / UAT — Issue #3974: typed database connector foundation

## Method

Generated the official `verify-work` prompt with:

```sh
scripts/gsd prompt verify-work 3974 --auto
```

The official command requires a numbered executed phase and an isolated GSD
runtime verifier. #3974 is an issue-local foundation slice, and this task's
canonical worker contract does not provide that role. Per the repository
delivery contract, the manual inline fallback was used and its coverage is
declared in `SUMMARY.md`.

## Automated deliverables

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Strict database definition/schema and defensive projections | `internal/connectors/database/database_test.go` | pass |
| Closed type planning, structured identity/catalog/read plan, bounded resources, cancellation | `internal/connectors/database/database_test.go` | pass |
| Shared descriptor/evidence admission separated from source execution | `internal/synccontract/native_admission_test.go` | pass |
| PostgreSQL reference seam and capability non-promotion | `internal/connectors/native/postgres/database_driver_test.go`, `internal/connectors/engine/database_definition_test.go` | pass |
| Adjacent CLI behavior remains intact | `go test -timeout 20m ./internal/cli -count=1` | pass |
| Static, generator, docs, boundary, smoke, and release gates | commands recorded in `VERIFICATION.md` | pass |

## Human-judgment routing

All deliverables are internal typed contracts with deterministic automated
proofs. No UI, product choice, credentialed connection, target mutation, or
human judgment is required for this slice.

## Verdict

Pass. The F1 foundation acceptance criteria are proven by automated evidence.
F2/F3/F4/F5/P-unit behaviors remain deliberately outstanding rather than being
claimed by this issue.
