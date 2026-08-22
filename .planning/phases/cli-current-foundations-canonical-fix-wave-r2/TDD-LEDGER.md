# TDD ledger — Canonical Foundation Fix Wave R2

## Assertion rule

Each row needs a behavior-first failure from the unmodified target, a focused Green after the smallest repair, and the aggregate regression named in the immutable review. Exit status alone is never credited; tests must observe a request, receipt, public projection, persisted state, or pre-I/O failure boundary.

| ID | Enforcement | Red evidence | Green evidence | Aggregate/regression |
| --- | --- | --- | --- | --- |
| FND-B01 | total source inventory arithmetic | pending: REST-only contradictory lock test | pending | `go test -timeout 20m ./cmd/connectorgen -count=1` |
| FND-B02 | REST descriptor disposition closure | pending: empty locked GET bundle | pending | `connectorgen validate`, preflight sweep |
| FND-B03 | GraphQL locked-root closure | pending: missing `Query.widgets` | pending | source/certification/GraphQL generator checks |
| FND-B04 | GitHub close/reopen reachability | pending: nested hook fails through `DryRunWrite` | pending | GitHub conformance/provider double |
| FND-B05 | Google Ads valid fixture witnesses | pending: invalid generic fixtures | pending | Google Ads conformance |
| FND-B06 | exhaustive proof visibility | pending: stale count hides row failure | pending | GitHub provider-double report |
| FND-B07 | schema-valid witness synthesis | pending: min/pattern/items reject | pending | provider-double capture |
| FND-B08 | GraphQL one-direction paging | pending: dual-direction root sends zero | pending | provider-double page/rate receipt |
| FND-B09 | schema-3 evidence closure | pending: decoder rejection | pending | `connectorgen evidence-gate` |
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
