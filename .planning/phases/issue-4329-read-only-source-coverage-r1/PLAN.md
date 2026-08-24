# Plan — issue 4329 source-cited read-only coverage

## Task Delivery Header

- Issue: Refs #4329 — allow source-cited operations to be declared intentionally read-only
- Base branch: main
- Merges into: main
- Delivery: PR open against `main`, with the API-reported base equal to `main`, all requested evidence committed, and full local `make verify` green.
- Working branch: fm/cli-read-only-source-coverage-r1
- Task: Add an explicit, declaration-owned non-mutating source-operation read-only refusal that remains distinct from missing-foundation gaps, cannot mask an executable route or a mutation, and is separately projected into operation evidence.
- Verification: Red/green behavioral tests through source projection and operation evidence; frozen GitHub artifact byte/hash measurements; full `make verify`; PR base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Explicit declared non-mutating source operation does not fail executable coverage | live | A real source-projection fixture accepts a cited GET only with the exact operation-ledger policy declaration and preserves a read-only result rather than a runtime gap. |
| An undeclared non-mutating operation still fails | live | Removing the declaration yields `source operation has no reachable executable operation` for the same source ID. |
| A read-only declaration cannot mask an executable route or a mutation | live | Adding an implemented GET command causes a contradictory-declaration finding; applying the marker to POST is rejected. |
| Read-only rows are separately reportable | live | The operation-evidence artifact has the declared read-only row and rollup while `missing_foundations` excludes it. |
| Sentry and Vercel mutations remain visible | live | The source-cited POST/PATCH operations with no action are enumerated for `cli-mutation-disposition-foundation-r1`, with no `read_only` declaration. |
| GitHub frozen artifacts do not drift | live | `wc -c` and `shasum -a 256` report the task's exact lock and descriptor values. |

## TDD execution slices

1. **Red:** Add behavioral tests that exercise source projection and the real operation-evidence path for declared, undeclared, and contradictory non-mutating source operations. Record the failure command and output.
2. **Green:** Add the smallest operation-ledger declaration reader/validator in source projection, integrate it into executable coverage, and ensure it refuses any executable counterpart or mutating source operation.
3. **Evidence:** Add the evidence-artifact row/rollup representation so a deliberate read-only refusal is neither a missing foundation nor a runtime/CLI gap. Enumerate rather than suppress the Sentry/Vercel mutation findings for the sibling foundation.
4. **Proof:** Run source-import/surface checks, artifact generation checks, exact GitHub byte/hash checks, and full `make verify`.
5. **Review:** Run inline `verify-work` and code review; record dispositions, merge current `origin/main` again before push, then open the main-targeted PR and read the base through the GitHub API.

## Skills and parity

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-naming`,
`golang-code-style`, `golang-lint`, and `golang-documentation`.

`connectorgen` is a developer generator command; this work adds no `pm`
command, connector user command, help topic, manual, or website page. Its
generated operation-evidence artifact is the applicable output surface.

## Authorized scope

The foundation uses `cmd/connectorgen/sourceprojection.go`, the required
evidence hook in `cmd/connectorgen/operationevidence.go`, and exactly one
closed `api_surface` operation-model enum member in the engine schema. No
Sentry or Vercel connector file is changed after the scope narrowed: their
mutations belong to the sibling disposition foundation.
