# TDD ledger — Canonical Foundation Fix Wave R2

## Assertion rule

Each row needs a behavior-first failure from the unmodified target, a focused Green after the smallest repair, and the aggregate regression named in the immutable review. Exit status alone is never credited; tests must observe a request, receipt, public projection, persisted state, or pre-I/O failure boundary.

| ID | Enforcement | Red evidence | Green evidence | Aggregate/regression |
| --- | --- | --- | --- | --- |
| FND-B01 | total source inventory arithmetic | `TestValidateSourceImportLockInventoryRejectsContradictoryRESTOnlyTotal` failed: `REST-only contradictory total was accepted`. | Same test passes after REST and total arithmetic are both checked before the GraphQL-free return. | `go test -timeout 20m ./cmd/connectorgen -count=1` passed (265.200s). |
| FND-B02 | REST descriptor disposition closure | `TestSourceProjectionRequiresReachableRESTReadOrConcreteGap` failed with no finding; `TestSourceProjectionCountsDeclaredPaginationParametersAsReachableInputs` failed when a declared pager was not credited; source validation reported 469 missing/incomplete locked reads. | Generic source route/field checks, source-bound partial CLI/API generation, and pager parameter accounting pass; cached GitHub regeneration validates with zero findings. | Package aggregate passed; `source-import github --check`, `surface-sync --check`, `connectorgen validate`, `certification-{candidates,sweep} --check`, and `TestEveryImplementedCommandPassesRuntimePreflight` passed. |
| FND-B03 | GraphQL locked-root closure | `TestSourceProjectionRequiresReachableGraphQLRootOrConcreteGap` failed with no finding for omitted `Query.widgets`. | Generic GraphQL identity/output/variables/command closure test passes; GitHub source validation has no GraphQL escape. | `go test -timeout 20m ./cmd/connectorgen -count=1` and the source/certification check-mode gates passed. |
| FND-B04 | GitHub close/reopen reachability | pending: nested hook fails through `DryRunWrite` | pending | GitHub conformance/provider double |
| FND-B05 | Google Ads valid fixture witnesses | pending: invalid generic fixtures | pending | Google Ads conformance |
| FND-B06 | exhaustive proof visibility | pending: stale count hides row failure | pending | GitHub provider-double report |
| FND-B07 | schema-valid witness synthesis | pending: min/pattern/items reject | pending | provider-double capture |
| FND-B08 | GraphQL one-direction paging | pending: dual-direction root sends zero | pending | provider-double page/rate receipt |
| FND-B09 | schema-3 evidence closure | pending: decoder rejection | Reserved for the final evidence-only closure commit: the gate requires HEAD's parent to be the completed code subject, so it cannot truthfully be completed before Groups 2–4. | `connectorgen evidence-gate` after final code and generated artifacts are green. |
| FND-B10 | ordinary data preservation | pending: SQS tags redact | pending | SQS/output suite |
| FND-B11 | cursor secret safety | pending: secret cursor public | pending | output/CLI page suite |
| FND-B12 | restriction fail-closed parse | pending: malformed restriction broadens request | pending | hook/config tests |
| FND-B13 | non-JSON receipt masking | pending: XML/text secret leak | pending | connectors/CLI output suite |
| FND-B14 | binary parameter authority | pending: undeclared `trace` reaches transport/file | pending | engine/runner binary suite |
| FND-B15 | PostgreSQL admission | pending: DB call before admission | pending | App/coordination/PostgreSQL suite |
| FND-B16 | Begin deadline/fence | pending: blocked Begin survives unit/lease | pending | synctransport/PostgreSQL suite |
| FND-B17 | continuation equality | pending: continuation-only change compares equal | pending | App/synccontract CAS suite |
| FND-B18 | full bulk acknowledgement | pending: partial nil-error terminalizes plan | pending | App reverse-finalization suite |
| FND-B19 | empty overwrite no-replay | pending: post-publish fault replays | pending | synctransport/App reconciliation |

## Execution log

- 2026-08-22: authoritative report read in full; final verdict and exact source identity independently verified. Planning begins before production edits.
- 2026-08-22: GSD adapter health, all required source lookups/prompts, and `agentcontractgen check` passed. Named-phase numeric incompatibility is recorded in `PLAN.md`; lifecycle runs inline with no role spawning.
- 2026-08-22: Group 1 Red — B01 contradictory REST-only source lock was accepted; B02/B03 missing generic REST and GraphQL routes produced no finding; a declared pagination parameter was not credited; an unbound implemented API command remained executable; and an incomplete declared read was not source-bound partial.
- 2026-08-22: Group 1 Green — generator tests now prove inventory arithmetic, REST/GraphQL closure, pager accounting, source-bound partial CLI/API projection, preservation of independent ETL stream coverage, and partial-command ledger validation. Cached locked GitHub import regenerated 614 exact source-bound execution gaps, 604 partial command dispositions, and 625 blocked endpoint ledgers; `go run ./cmd/connectorgen validate internal/connectors/defs` reports zero findings.
- 2026-08-22: Group 1 aggregate Green — `go test -timeout 20m ./cmd/connectorgen -count=1` passed in 265.200s. The cached source lock, surface-sync, certification-candidates, certification-sweep, validator, and `TestEveryImplementedCommandPassesRuntimePreflight` all passed in check mode. The GitHub partial rows now require their named locked source operation; old provider-refusal observations are excluded if current declaration availability is partial.
