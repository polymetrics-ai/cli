# #3987 — four-path warehouse conformance context

## Task Delivery Header

- Issue: Refs #3987 — Postgres Parity: prove the warehouse-mediated flow and mode matrix.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with the stated local verification recorded and its API-reported base read back.
- Pull request: [#4195](https://github.com/polymetrics-ai/cli/pull/4195) (`test(certification): prove four warehouse flow contracts`). After opening, `gh-axi pr list --state open --base integration/4015-mvp-flat-r1 --head fm/cli-3987-four-path-conformance-r1` returned exactly #4195, verifying GitHub's reported base selection through the required API-backed wrapper query.
- Working branch: `fm/cli-3987-four-path-conformance-r1`.
- Task: Add the missing conformance-matrix proof for the four canonical GitHub/PostgreSQL warehouse-mediated directions, preserving existing route behavior and making each direction, current executable mode, and the change-capture restriction independently observable.
- Verification: Focused conformance tests; `go test -timeout 20m` for changed packages and `internal/cli`; generated-certification and connector-boundary checks; build, docs validation, smoke, lint, website docs generation twice, and the repository’s remaining individual local gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Four canonical directions are distinct contracts | fake | Deterministic production-composition tests use bounded provider/database fixtures. Existing opt-in fresh-binary route proofs remain the live evidence; the new matrix makes their identities and warehouse boundary assertions independently fail-closed in CI. |
| Warehouse mediation is enforced and observable | fake | The production composition is exercised with a stage that records source owner, sealed receipt/workset, target input, acknowledgement, and checkpoint order; no live credential is required or used. |
| Closed mode coverage reflects the current branch | fake | Tests derive `synccontract.AllModes()` and verify current descriptor/preflight admission rather than use stale issue text. |
| `change_capture` is source-only PostgreSQL workset input | fake | A production-composition refusal test proves it cannot resolve as a target mode and a successful source-workset path asserts the derived workset boundary. |
| A direction-specific regression fails | fake | The TDD ledger records a scratch, post-schema-compilation break of one direction and the exact focused test failure, followed by restoration. |

## Authoritative branch facts

- The branch began at `ff6a87101` on `origin/integration/4015-mvp-flat-r1`; the default checkout was corrected before planning and before any production edit.
- The closed vocabulary is `synccontract.AllModes()`: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`.
- The #3987 instruction to reject `incremental_dedupe_history` is stale. PostgreSQL’s native route and GitHub’s dedupe/history source behavior were made executable by merged PRs #4187 and #4188. This task must prove current behavior, not restore the old refusal.
- `change_capture` remains a PostgreSQL source contract. The PostgreSQL source/destination transport declarations intentionally omit it from the normal target-mode intersection; it is delivered through the derived warehouse workset path.

## Route proofs versus this conformance gate

| Direction | Existing route proof | Gap owned here |
| --- | --- | --- |
| API → API | #4185: GitHub issues through durable warehouse to issue labels, with independent GitHub read-back. | A named matrix contract that prevents its evidence from standing in for the other three directions. |
| API → database | Existing fresh-binary authenticated GitHub → PostgreSQL tests, including the 90k-commit regression in `internal/cli/postgres_transport_binary_integration_test.go`. | Matrix-level ownership, warehouse invariant, and current closed-mode accounting. |
| database → API | #4186: PostgreSQL polling watermark through the warehouse to GitHub labels, with provider read-back. | Same matrix-level non-substitutability and explicit target/source boundary proof. |
| database → database | #4184 and the PostgreSQL binary integration path: source, durable warehouse, managed target, read-back, and checkpoint. | Same matrix-level non-substitutability and explicit target/source boundary proof. |

## Scope guard

This is a cross-flow certification conformance gate, not a connector implementation lane. It must not change a connector definition, generic transport registration, PostgreSQL certification profile/adapter, GitHub direct-read candidates/output contract, CLI help/dispatch, final #3978 live certification, or `internal/connectors/certify/stages_source.go` / `allStagesPassed`.
