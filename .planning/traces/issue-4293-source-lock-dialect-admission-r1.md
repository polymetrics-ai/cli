# Plan — Issue #4293 source-lock dialect admission R1

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base branch: `codex/4293-mapping-media-wildcards-r1` at independently reviewed remote commit `d18f6b17315316659af773b2ccccf4275918c0b6`.
- Merges into: `codex/4293-mapping-media-wildcards-r1` → `fm/cli-top100-declaration-batch-r1` → `main`; integration and merge remain coordinator/captain decisions.
- Delivery: a scoped commit and pushed branch, with a #4293 evidence comment requesting independent review. This task does not claim merge readiness.
- Working branch: `fix/4293-source-lock-dialect-admission-r1`.
- Task: retain only the documented Batch R1 legacy v2 source-lock dialect facts (`source_operation` and GitLab `rest.path_bridge`) through strict source import and declaration/mapping reconciliation.
- Verification: focused `cmd/connectorgen` import/mapping/cohort tests; `gofmt`; `go vet`; `agentcontractgen check`; `git diff --check`; and the non-writing Batch R1 ten-lock cohort check when feasible.

## Scope and design

The reviewed base contains the intended narrow source-import model, and this
slice closes one corresponding declaration-admission gap:

1. `source_operation` is a typed, object-only enrichment whose provider-owned JSON fragment is retained as raw evidence. The enclosing source-lock wire remains strict. Declaration admission now reuses that exact type rather than accepting an unchecked `json.RawMessage`.
2. `rest.path_bridge` is a typed two-prefix mapping rule. It is validated before import and again while declaration-admission reconstructs reviewed source identities.
3. The mapping manifest continues to require exact source URL and source-location citations for every operation and fact.

This slice must not broaden unknown-field handling, alter source locks or connector definitions, or touch execution, transport, certification, or runtime behavior. It adds regression coverage around the real Batch R1 locks and the existing closed error boundaries instead of inventing a new generic dialect decoder.

## Acceptance evidence

| Requirement | Evidence to add or run |
| --- | --- |
| Sentry, Docker Hub, and Notion retained `source_operation` facts survive | Parse immutable v2 locks, assert every REST row retains a non-empty provider object, assert a known provider `summary`, and reconcile the same identity through declaration admission. |
| GitLab `rest.path_bridge` survives and is usable | Parse the immutable lock, assert its exact prefixes, reconstruct a canonical source document from its retained source-operation evidence, and assert declaration admission maps `getApiV4Projects` from `/api/v4/projects` to `/projects`. |
| A present bridge is validated | Feed an invalid in-memory bridge before import; reject it before any fetch occurs. |
| Parsing stays closed | Mutate only an in-memory JSON copy with misspelled lock/bridge fields; strict source import and declaration-admission parsing must reject both. Provider fields inside the documented `source_operation` object remain intentionally opaque provider grammar. |
| Citations/backlinks remain required | Use a reconciled GitLab operation to prove an exact source citation passes and a missing or drifted citation fails; retain the existing source-operation-mapping manifest checks as the full-schema proof. |

## TDD record

The historical red condition is visible in pre-admission source-import code: a strict legacy decoder rejected v2 locks with `json: unknown field "source_operation"` and without the typed path-bridge wire rejected GitLab `rest.path_bridge`. The new real-lock tests were written before production changes. They found one current mapping-side inconsistency: source import rejected a malformed `source_operation: []`, but declaration admission accepted it through an unchecked `json.RawMessage` leaf.

The red/green/refactor sequence is:

1. Add the real-lock compatibility and strict-boundary tests.
2. Run focused tests. The new malformed-provider test was red: `parseDeclarationAdmissionSourceLock` returned `nil` error for `source_operation: []`.
3. Replace only the two declaration-admission operation-wire raw leaves with the existing object-only `sourceImportOperationEnrichment` type, then rerun the test green.
4. Refactor only for concise deterministic test helpers; do not make a generic `map[string]any` source-lock escape hatch.
5. Run focused import, mapping, cohort, vet, agent-contract, and diff checks; record any unrelated baseline limitation separately.

## Green evidence

The initial real-lock tests established the existing source-import and bridge
behavior. Adding the mapping-side non-object `source_operation` case then
produced the required red result:

```text
declaration-admission dialect error = <nil>, want source_operation must be an object
```

The repair changes only both declaration-admission operation wires to reuse
the importer-owned `sourceImportOperationEnrichment`; no generic unknown-field
handling was added. The focused green test then passed in 14.176s:

```text
GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen -run '^(TestBatchOneSourceOperationFactsRemainStrictlyAdmittedAndReconciled|TestBatchOneGitLabPathBridgeAndSourceOperationStayClosed)$' -count=1 -v
```

The Sentry, Docker Hub, and Notion subtests each
verify all retained provider objects plus one known summary and a reconciled
source identity. The GitLab test verifies canonical source-document assembly,
the `/api/v4` → connector-relative bridge, exact citation retention, invalid
bridge pre-fetch refusal, and strict rejection of misspelled REST and bridge
fields.

The broader focused import/mapping/cohort suite then passed in 24.401s:

```text
GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen -run '^(TestBatchOneSourceOperationFactsRemainStrictlyAdmittedAndReconciled|TestBatchOneGitLabPathBridgeAndSourceOperationStayClosed|TestSourceImportGitLabCanonicalEvidenceProjectsLockedInventoryWithoutFetcher|TestParseSourceImportLockProviderFragmentsKeepRootClosed|TestSourceOperationMappingCheckRejectsUncitedOrContradictoryBinaryMedia|TestSourceOperationMappingCheckRejectsUncitedAndContradictoryApplicabilityFacts|TestBatch1SourceOperationMappingCohortCheckAcceptsTrackedDenominator)$' -count=1 -v
```

It includes the ten-lock / 4,341-source-operation cohort check, the complete
GitLab canonical-evidence import without provider fetch, and cited mapping
fact rejection cases. The following mechanical gates also passed after the
production repair:

```text
GOFLAGS='-p=3' go vet ./cmd/connectorgen
GOFLAGS='-p=3' go run ./cmd/agentcontractgen check
GOFLAGS='-p=3' go run ./cmd/connectorgen source-operation-mapping-cohort data/connector-canon/batch1-source-operation-mapping-cohort.json --check
git diff --check
```

The cohort command reports exactly `10 connector(s), 4341 source operation(s),
0 finding(s)`; agentcontract reports canonical contract/projections current.

## Lifecycle and skills

- CodeGraph was checked before locating code. This worktree has no `.codegraph/` directory, so no CodeGraph query was available.
- `scripts/gsd doctor` and the generated lifecycle prompts (`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`) were run. A Pi role worker is not available in this isolated Codex task, so the lifecycle is recorded and executed inline.
- Read and applied: `go-engineering`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## Residual boundary

This parser/admission slice does not make an operation executable, certify credentials, alter a lane disposition, or bypass source evidence. A source lock that has no canonical retained source contract continues to use its pinned provider artifact path; that is intentional and outside this scope.
