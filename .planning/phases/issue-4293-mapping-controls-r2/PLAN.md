# Plan — Issue #4293 mapping controls R2

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator
- Base branch: `fix/4293-source-operation-multilane-manifest-r1` at immutable commit `27608b31ed0f3b138fe6218188ca02a084b4d8eb`
- Merges into: `fix/4293-source-operation-multilane-manifest-r1` → its parent integration branch → `main`; no merge or pull request is authorized for this task.
- Delivery: A scoped branch committed and pushed with mapping-control checks green, plus a #4293 proof comment.
- Working branch: `codex/4293-mapping-controls-r1`
- Task: Repair MAP-001/002/003 by adding a fixed Batch R1 ten-lock cohort anchor, source-backed mutation cell admission, and exact source-node-bound applicability facts. Connector-local lane matrices remain referenced inputs; their operation rows are not copied into the cohort artifact.
- Verification: Focused `cmd/connectorgen` red/green tests; canonical cohort `--check`; JSON schema parsing; `gofmt`; `go vet` for changed packages; `go run ./cmd/agentcontractgen check`; and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Batch R1 retains exactly ten locks and 4,341 source rows | live | The repository-only cohort check parses all ten current locks, compares each lock and sorted source-ID digest/count to the tracked anchor, and reports 4,341 rows. |
| A provider mutation retains both write lanes | live | Focused fixtures reject missing `direct_write` and `reverse_etl` cells for REST mutations and a GraphQL `Mutation.*` root; a source-cited non-mutating POST remains a negative boundary. |
| ETL, binary, and sync facts cannot be self-asserted | live | Focused fixtures reject omitted record shape, a fact cited to another source node, and facts whose applicability contradicts the cited pagination/media/event evidence. |
| Connector-local matrices remain one source of lane rows | live | The cohort anchor records canonical connector-local matrix input paths but contains no source-operation or cell rows; the checker validates lock/identity facts only. |

## Discuss and planning result

The issue, frozen review ledger, current source locks, and in-flight Stripe/Vercel matrices remove the remaining design ambiguity:

1. The cohort is a source-lock denominator anchor, not a second matrix. It stores exact lock bytes digests, sorted source-ID digests, per-lock counts, the 10 fixed connector identities, the 4,341 total, and canonical connector-local matrix input paths. It does not reproduce any matrix operation or lane cell.
2. Mapping facts are a strict cited sidecar. Their citations must resolve to the exact locked operation source node (same source URL and same locked location), rather than merely the same document URL. Record shape and explicit ETL/binary/sync applicability join the existing pagination, scope, media, and event/cursor facts.
3. REST `PUT`, `PATCH`, and `DELETE`, and GraphQL `Mutation.*`, are source identities that must state `mutation` and independently retain both write-lane dispositions. A REST `POST` is not inferred as a mutation: a source-node-cited mutation fact makes the distinction and is tested with a non-mutating POST boundary.

## TDD slices

1. Add focused failing tests and fixture mutations for the Batch R1 cohort digest/count/membership checks, all mutation boundaries, record shape, wrong-node citations, and ETL/binary/sync contradictions. Record the red command before implementation.
2. Add the closed cohort schema and check-only authoring command. Anchor the ten current source locks and their source-ID digests in a tracked canonical JSON artifact, with opaque connector-local matrix input paths.
3. Extend the existing mapping schema/model and validator with node-bound record-shape, mutation, and lane applicability facts. Enforce required write cells and source-fact/cell consistency without calling any runtime, credential, transport, certification, or provider-I/O surface.
4. Refactor only for deterministic diagnostics and concise helpers. Run the scoped green gates and record broader baseline work separately.

## GSD and skill record

- Generated and executed inline/manual fallbacks for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` via `scripts/gsd prompt`; Pi-compatible isolated workflow agents are not available in this task environment, and no delegation is authorized.
- `scripts/gsd doctor`, command-source resolution, and `go run ./cmd/agentcontractgen check` passed before production edits.
- Skills used: `connector-lane-build-order`, `go-engineering`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

## Scope and safety boundary

- Included: `cmd/connectorgen` mapping controls, embedded mapping schemas, focused fixtures/tests, a canonical cohort manifest, and this planning evidence.
- Excluded: source-lock bytes, connector definitions and artifacts, provider imports, runtime/engine execution, credentials, transports, certification, and generated output rewrites.
- Atlas: no runtime foundation is proposed. The existing source-projection authoring seam is reused only; no Atlas catalog change or approval is required.
- Developer `connectorgen` help is covered by focused tests. No `pm` runtime command, manual, website page, completion, credential path, or JSON-output surface changes, so those parity surfaces are not applicable.
