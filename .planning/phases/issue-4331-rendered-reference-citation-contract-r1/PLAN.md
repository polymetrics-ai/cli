# Issue 4331 — Rendered-reference citation contract

## Task Delivery Header

- Issue: Refs #4331 — fix(connectorgen): support rendered-reference citations in v3 source locks
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` with its checks green.
- Working branch: fm/cli-rendered-reference-citation-contract-r1
- Task: Extend the v3 source-lock REST document contract with optional/defaulted document kinds (`openapi`, `rendered_reference`, `bundle`, and explicit `unavailable`), integrity-provenance requirements, coverage confidence, and operation citations without changing legacy or OpenAPI contract behavior.
- Verification: TDD red/green importer-projection tests; `go run ./cmd/connectorgen validate`; targeted `go test -timeout 20m ./cmd/connectorgen`; generation and repository validation gates listed in `VERIFICATION.md`; manual code review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A rendered capture imports and projects through the real pipeline | live | The real importer emits a projected REST operation from an `application/json` rendered capture with the cited route and source document, which is absent before the feature. |
| Mixed OpenAPI and rendered documents retain both operation paths | live | A real mixed lock projects both distinct operations and retains its OpenAPI version inventory. |
| Captured rendered bytes cannot be altered | live | The real lock validator rejects a fixture whose bytes fail its declared SHA-256. |
| Every rendered operation has a same-origin absolute citation | live | The real lock validator rejects missing and foreign-origin citations. |
| Existing OpenAPI v3 projection remains stable | live | A real OpenAPI-only fixture produces byte-identical projection JSON before and after a rendered-document fixture is added. |
| Schema 1 and schema 2 remain supported | live | Existing legacy source-lock fixtures validate through the real importer. |
| Unavailable source inventory remains visibly blocked | live | The real source-projection validator returns a source-traced blocking finding rather than accepting an empty inventory. |

## Execution plan

1. **Discuss / plan (inline fallback).** The Pi runtime is unavailable in this environment, so execute the GSD prompts inline and record the decision, scope, skills, and TDD sequence in this phase directory.
2. **Red.** Reproduce batch 6/7 source-projection failures with `go run ./cmd/connectorgen validate`, then add behavioral tests using real source-lock fixtures and run them before production edits.
3. **Green.** Add the discriminated source document type and strict same-origin citation validation in `cmd/connectorgen/sourceimport.go`; preserve shared importer/projection identity and count validation.
4. **Refactor / regression.** Keep OpenAPI validation path explicit and unchanged; add test helpers only where they make captured-document fixtures auditable.
5. **Verify / review.** Run the scoped tests and non-test repository gates individually, inspect the diff for the no-fetch invariant, and execute the generated GSD verification and review workflow inline.

## Constraints and decisions

- The kind is discriminated and optional. Absent `kind` is `openapi`, preserving the landed v3 wire projection; `rendered_reference` requires captured bytes, content type, coverage confidence, and per-operation citations, and never contributes an OpenAPI version.
- The authorized scope extension covers 50 connectors: generic rendered captures, archive bundles (hash-verified and enumerated from the lock without archive extraction), confidence declarations, and explicit source-traced unavailable gaps. It intentionally does not accept the colliding batch 8/9/10 schema-3 field names.
- A citation URL is provenance only. Source import must never fetch it; captured bytes plus SHA-256 are the evidence.
- This is shared foundation work only: do not modify the batch 6/7 connector locks or engine code.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-documentation`.
- CLI help/manual/website parity is not applicable: this changes the build-time `connectorgen` source-lock contract and does not change the `pm` command surface.
