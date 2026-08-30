# Plan — CircleCI source-lane matrix R1

## Task Delivery Header

- Issue: Refs #4382 — CircleCI — source-to-seven-lane matrix
- Base branch: `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`
- Delivery: Scoped branch committed and pushed with focused checks green; no pull request is opened by this task.
- Working branch: `fix/4382-circleci-track-a-r1`
- Task: Materialize a cited CircleCI source-to-seven-lane matrix for every retained lock row, add local adversarial validation, and preserve mapping-only planning evidence without altering execution behavior.
- Verification: Focused local matrix test, red/green ledger, JSON parse/count checks, `gofmt`, diff checks, relevant admission/definition checks recorded faithfully, and manual inline review evidence.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every retained source row remains visible | live | The validator compares the exact pinned lock ID set to matrix rows; deleting one in memory fails with `source row absent from matrix`. |
| Every row has seven cited cells | live | The validator requires each named lane exactly once and asserts 777 total cells. |
| Paging and mutation candidates are explicit | live | Adversarial probes change a cursor ETL cell or mutation reverse-ETL cell and the same validator rejects the matrix. |
| Artifacts cannot create source IDs/cells | live | The validator resolves each artifact record and link against retained artifacts and the matrix; an unknown source backlink fails. |
| No runtime semantics are asserted | live | The validator rejects `implemented` and `missing_foundation` states in this Track A matrix. |

## Scope guard

Allowed files are CircleCI definition source/matrix/test material and this phase's
planning evidence. Do not change shared engine, generator, runtime, Foundation Atlas,
credential handling, provider transport, connector execution, or connector-specific
artifacts other than focused matrix/test evidence.

## Skills used

- `connector-lane-build-order`: source-lock authority, lane classification and Atlas
  gate.
- `go-engineering` (fundamentals, testing, error handling, safety, data structures):
  local pure validation and table-driven adversarial tests.
- `agentic-ETL` guidance: preserve ETL as evidence classification; do not claim a
  runnable pipeline from a pageable response.

## TDD plan

1. Add a local CircleCI matrix validator test before the matrix exists. It must load
   the pinned lock, require exact one-to-one source operation accounting and seven
   cells, then fail because the materialized matrix is absent.
2. Add adversarial tests that mutate an otherwise valid matrix to prove a hidden source
   row and an invalid artifact backlink are rejected. Include paging/ETL and mutation
   cell independence probes.
3. Materialize the matrix from the pinned lock, with per-operation source facts and
   citations. Classify cells only from retained source data.
4. Run focused Go tests, gofmt, JSON parsing, source-lock/matrix validation, and the
   project declaration admission check. Record any unrelated broad-suite baseline
   failures without widening scope.

## Acceptance checks

| Requirement | Evidence |
| --- | --- |
| All 111 lock rows represented | local matrix reconciliation test checks exact ID sets |
| Seven cells per row | same validator requires the complete lane set exactly once |
| 9 cursor candidates explicitly dispositioned | test pins IDs and requires ETL + sync `mapped_unproven` |
| 50 mutation candidates explicitly dispositioned | test pins IDs and requires independent direct-write + reverse-ETL cells |
| No secret/identity semantic loss | focused assertions preserve declared webhook field names and `project-slug` facts |
| Artifact links cannot invent source IDs | backlink validator resolves every link to an extant cell |
| No execution implication | all candidate states are `mapped_unproven`; no `implemented` or `missing_foundation` cells |
