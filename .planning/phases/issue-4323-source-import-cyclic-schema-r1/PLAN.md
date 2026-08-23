# Plan — issue 4323 retain cyclic schema references as source-bound gaps

## Task Delivery Header

- Issue: Refs #4323 — retain cyclic schema references as source-bound gaps
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, committed on
  `fm/cli-source-import-cyclic-schema-r1`, with full local `make verify` green
  and GitHub API base read-back equal to `main`.
- Working branch: fm/cli-source-import-cyclic-schema-r1
- Task: Make the shared source importer preserve recursive schema references as
  existing missing-foundation gap evidence with schema/pointer provenance,
  without losing operations or changing the v3 source-lock/provenance model.
- Verification: Red and green real-import-path Go tests, an affected connector
  import, frozen GitHub artifact hash/byte measurements, generated checks, full
  `make verify`, diff review, and PR base/review-route checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A direct self-reference imports and records a source-bound gap | live | The real importer output retains the operation and contains a missing-foundation entry naming the self schema and its source pointer. |
| Mutual recursion and a nested cycle do not disappear or expand | live | Real importer fixtures retain the affected operations and report their reference cycles at the schema pointer reached through nesting. |
| A non-cyclic schema remains unaffected | live | The real importer output contains no cycle-derived missing-foundation entry. |
| A real affected connector succeeds | live | `source-import <connector> --check` accepts the imported descriptor and its source trace; no credentials are used. |
| GitHub frozen artifacts remain byte-identical | live | `wc -c` and `shasum -a 256` report the fixed source-lock and descriptor values from the task brief. |

## TDD execution slices

1. **Red:** add real importer-path tests for direct, mutual, deeply nested, and
   non-cyclic schemas. Run the narrow target and record the pre-change grammar
   rejection.
2. **Green:** make the smallest importer change that transforms a reference
   cycle into existing missing-foundation evidence carrying source trace and
   canonical mapping; re-run the target tests and inspect exact gap fields.
3. **Proof:** run an affected real connector import, frozen GitHub artifact
   measurements, generated checks, and the required complete `make verify`.
4. **Review:** inspect the diff for v3/provenance regression, run inline code
   review under the GSD fallback, and record dispositions before opening the PR.

## CLI docs parity

`connectorgen source-import` is an internal generator command and does not
change `pm`, a connector command surface, help, docs, the website, completion,
or the generated manual. Those parity surfaces are not applicable.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-lint`.

Resolved command path: `scripts/gsd doctor`; `scripts/gsd sources` and
`scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. The generated prompts are executed inline
because this direct-PR worker has no compatible isolated GSD role runtime.
